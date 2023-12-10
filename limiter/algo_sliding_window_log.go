package limiter

/*
Algorithm: Sliding Window Log
Use a heap to store the requests.
* Push incoming requests in the heap, remove entries older than window
  size from the heap
* Rate limit if the number of requests in the heap is above allowed rate
*/

import (
	"container/heap"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"
)

var (
	ErrLimitExceeded = fmt.Errorf("rate limit exceeded")
)

// SlidingWindowLogRateLimiter satisfies the RateLimiter interface
// This is used to support rate limiting per client (or other criterion)
type SlidingWindowLogRateLimiter struct {
	config *SlidingWindowLogConfig
	swMap  map[string]*slidingWindowLog
}

// Allow returns nil if the request is allowed, otherwise returns an error
func (s *SlidingWindowLogRateLimiter) Allow(id string) error {
	swl := s.swMap[id]
	if swl == nil {
		swl = &slidingWindowLog{
			Id:              id,
			windowSize:      s.config.windowLen,
			maxRequestCount: s.config.requestPerSec * int(s.config.windowLen.Seconds()),
			requestHeap:     newRequestHeap(),
			mu:              &sync.Mutex{},
		}
		s.swMap[id] = swl
	}
	return swl.allowRequest()
}

func (s *SlidingWindowLogRateLimiter) GetLimit() int {
	return s.config.requestPerSec * int(s.config.windowLen.Seconds())
}

func (s *SlidingWindowLogRateLimiter) Unregister(s2 string) { delete(s.swMap, s2) }

func (s *SlidingWindowLogRateLimiter) Stop() {}

func (s *SlidingWindowLogRateLimiter) Stats() interface{} {
	data := make(map[string]interface{})
	for key, fwc := range s.swMap {
		data[key] = map[string]interface{}{
			"request_count": fwc.requestHeap.Len(),
			"capacity":      fwc.maxRequestCount,
			"interval":      fwc.windowSize,
		}
	}
	return data
}

// requestLog is an entry in the sliding window log
type requestLog struct {
	timestamp time.Time
	request   interface{}
}

// requestHeap is used to store request logs
type requestHeap struct {
	requests []requestLog
}

func newRequestHeap() *requestHeap {
	return &requestHeap{
		requests: []requestLog{},
	}
}

func (r *requestHeap) Peek() requestLog { return r.requests[0] }

func (r *requestHeap) Push(x interface{}) {
	r.requests = append(r.requests, x.(requestLog))
}

func (r *requestHeap) Pop() interface{} {
	n := len(r.requests)
	x := r.requests[n-1]
	r.requests = r.requests[:n-1]
	return x
}

func (r *requestHeap) Len() int {
	return len(r.requests)
}

func (r *requestHeap) Less(i, j int) bool {
	return r.requests[i].timestamp.Before(r.requests[j].timestamp)
}

func (r *requestHeap) Swap(i, j int) {
	r.requests[i], r.requests[j] = r.requests[j], r.requests[i]
}

// check heap and sort interfaces
var _ heap.Interface = &requestHeap{}
var _ sort.Interface = &requestHeap{}

// SlidingWindowLogConfig stores the configuration after parsing arguments
type SlidingWindowLogConfig struct {
	windowLen     time.Duration
	requestPerSec int
}

// Parse parses the args and stores the values in SlidingWindowLogConfig
func (swlc *SlidingWindowLogConfig) Parse(config RateConfig) error {
	limit, err := strconv.ParseInt(config["request_per_sec"], 10, 32)
	if err != nil {
		return err
	}
	swlc.requestPerSec = int(limit)

	swlc.windowLen, err = time.ParseDuration(config["window_size"])
	if err != nil {
		return err
	}
	return nil
}

// slidingWindowLog is a sliding window log implementation for rate limiting
// for each identity/client (user, ip, etc.)
type slidingWindowLog struct {
	mu *sync.Mutex
	// Id is the id of the user or IP address
	Id string
	// maxRequestCount is the maximum number of requests allowed in the window
	maxRequestCount int
	// windowSize is the size of the sliding window
	windowSize time.Duration
	// heap timestamps is the log of requests
	*requestHeap
}

// cleanup removes requests that fall out of the window
func (s *slidingWindowLog) cleanup() {
	startTime := time.Now().Add(-s.windowSize)
	for s.Len() > 0 {
		if s.Peek().timestamp.After(startTime) {
			break
		}
		// remove the requests outside the window
		s.Pop()
	}
}

// allowRequest returns an error if rate limit is reached otherwise nil
func (s *slidingWindowLog) allowRequest() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// submit the request
	s.requestHeap.Push(requestLog{
		timestamp: time.Now(),
	})
	s.cleanup()
	if s.Len() > s.maxRequestCount {
		return ErrLimitExceeded
	}
	return nil
}
