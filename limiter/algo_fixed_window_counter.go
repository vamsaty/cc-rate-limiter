package limiter

/*
Algorithm: Fixed Window Counter
Implements the Fixed Window Counter algorithm to rate limit requests.
If the time elapsed between two consecutive requests is more than the
window's length, the window is reset. Otherwise, rate limit if the window
is full.
*/

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

var (
	ErrWindowFull = errors.New("window is full")
)

// WindowLimiterImpl is a fixed window counter implementation for rate limiting.
// This supports rate limiting per client (or any other criterion)
type WindowLimiterImpl struct {
	// windows is a map of windowId to window
	*sync.Mutex
	windowMap map[string]*window
	config    *WindowConfig
}

// Allow checks if a request can be allowed
func (w *WindowLimiterImpl) Allow(id string) error {
	w.Lock()
	defer w.Unlock()

	fwc := w.windowMap[id]
	if fwc == nil {
		fwc = &window{
			Id:              id,
			windowSize:      w.config.WindowSize,
			maxRequestCount: w.config.MaxRequestCount,
			requestCount:    0,
			startTime:       time.Now(),
		}
		w.windowMap[id] = fwc
	}
	return fwc.allowRequest()
}

// Unregister removes the window for the given id
func (w *WindowLimiterImpl) Unregister(s string) {
	w.Lock()
	defer w.Unlock()
	delete(w.windowMap, s)
}

func (w *WindowLimiterImpl) Stop() {}

// Stats returns the stats for the window limiter
func (w *WindowLimiterImpl) Stats() interface{} {
	data := make(map[string]interface{})
	for key, fwc := range w.windowMap {
		data[key] = map[string]interface{}{
			"size":     fwc.requestCount,
			"capacity": fwc.maxRequestCount,
			"interval": fwc.windowSize,
		}
	}
	return data
}

func (w *WindowLimiterImpl) GetLimit() int { return w.config.MaxRequestCount }

// WindowConfig is the configuration for a window
type WindowConfig struct {
	// WindowSize is the size of the window (interval)
	WindowSize time.Duration
	// MaxRequestCount is the maximum number of requests allowed in a window
	MaxRequestCount int
}

// Parse parses the args and populates the WindowConfig
func (wc *WindowConfig) Parse(config map[string]string) error {
	var err error
	var value int64

	// parse WindowSize
	wc.WindowSize, err = time.ParseDuration(config["window_size"])
	if err != nil {
		panic(err)
	}

	// parse MaxRequestCount
	value, err = strconv.ParseInt(config["max_request_count"], 10, 64)
	if err != nil {
		panic(err)
	}
	wc.MaxRequestCount = int(value)

	return nil
}

// window represents the fixed window counter (per identity - user, ip, etc.)
type window struct {
	// Id is used to identify the entity for which the window is created
	Id string
	// requestCount is the total number of requests in the window
	requestCount int
	// maxRequestCount specifies the max number of requests allowed in the window
	maxRequestCount int
	// windowSize specifies the window's length
	windowSize time.Duration
	// startTime specifies the start time of the window
	startTime time.Time
}

// reset resets the request count and start time
func (w *window) reset() {
	if time.Now().Sub(w.startTime) > w.windowSize {
		w.requestCount = 0
		w.startTime = time.Now()
	}
}

// allowRequest checks if a request can be allowed
func (w *window) allowRequest() error {
	// check window reset
	w.reset()

	// check if window is full
	if w.requestCount >= w.maxRequestCount {
		return ErrWindowFull
	}
	w.requestCount++
	return nil
}
