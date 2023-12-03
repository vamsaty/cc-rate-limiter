package limiter

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

var (
	ErrBucketEmpty = fmt.Errorf("bucket is empty")
	//ErrBucketFull  = fmt.Errorf("bucket is full")
)

// TokenBucketConfig is the configuration for the token bucket
type TokenBucketConfig struct {
	Capacity   int     // max number of tokens in the bucket
	RefillRate float64 // tokens pushed into the bucket per second
}

func (tbc *TokenBucketConfig) Parse(config map[string]string) error {
	var err error
	var value int64

	value, err = strconv.ParseInt(config["capacity"], 10, 32)
	if err != nil {
		return err
	}
	tbc.Capacity = int(value)

	tbc.RefillRate, err = strconv.ParseFloat(config["refill_rate"], 64)
	if err != nil {
		return err
	}
	return nil
}

// tokenBucket is a token bucket implementation for rate limiting
type tokenBucket struct {
	mu         *sync.Mutex
	Id         string    // Id of the bucket - username/userid or IP address
	tokens     int       // current number of tokens in the bucket
	capacity   int       // max number of tokens in the bucket
	refillRate float64   // tokenPushRate is the number of tokens pushed into the bucket per second
	lastRefill time.Time // timestamp of last request
}

// Stop stops the token pusher
func (tb *tokenBucket) Stop() {}

// allowRequest checks if a request can be allowed
func (tb *tokenBucket) allowRequest() error {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens <= 0 {
		return ErrBucketEmpty
	}
	tb.tokens--
	return nil
}

func (tb *tokenBucket) nextRefillSize() int {
	elapsed := time.Since(tb.lastRefill)
	tokens := int(elapsed.Seconds()*tb.refillRate) + tb.tokens
	if tokens > tb.capacity {
		tokens = tb.capacity
	}
	return tokens
}

func (tb *tokenBucket) refill() {
	elapsed := time.Since(tb.lastRefill)
	tb.tokens += int(elapsed.Seconds() * tb.refillRate)
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastRefill = time.Now()
}

func newTokenBucket(Id string, config *TokenBucketConfig) *tokenBucket {
	tb := tokenBucket{
		Id:         Id,
		tokens:     config.Capacity, // initially bucket is full
		capacity:   config.Capacity,
		refillRate: config.RefillRate,
		lastRefill: time.Now(),
		mu:         &sync.Mutex{},
	}
	return &tb
}

// TBLimiter is a token bucket limiter, satisfying the RateLimiter interface
type TBLimiter struct {
	// bucketMap is a map of userId to token bucket. Generally a bucket is
	// created for each user/IP address
	bucketMap map[string]*tokenBucket
	// Mutex is a lock for bucketMap
	*sync.RWMutex
	// shutDown is a channel to stop the token pusher
	shutDown chan struct{}
	// config is the configuration for the token bucket
	config *TokenBucketConfig
}

// Allow checks if a request can be allowed.
func (tbl *TBLimiter) Allow(Id string) error {
	tbl.Lock()
	defer tbl.Unlock()

	bucket := tbl.bucketMap[Id]
	if bucket == nil {
		bucket = newTokenBucket(Id, tbl.config)
		tbl.bucketMap[Id] = bucket
	}
	return bucket.allowRequest()
}

// Unregister remove a bucket from the @bucketMap
func (tbl *TBLimiter) Unregister(Id string) {
	tbl.Lock()
	defer tbl.Unlock()

	if bucket := tbl.bucketMap[Id]; bucket != nil {
		bucket.Stop()
		delete(tbl.bucketMap, Id)
	}
}

// Stop stops the token pusher for all buckets
func (tbl *TBLimiter) Stop() {
	for _, bucket := range tbl.bucketMap {
		bucket.Stop()
	}
}

func (tbl *TBLimiter) Stats() interface{} {
	data := make(map[string]interface{})
	for key, bucket := range tbl.bucketMap {
		data[key] = map[string]interface{}{
			"tokens":         bucket.tokens,
			"capacity":       bucket.capacity,
			"refill_rate":    bucket.refillRate,
			"current_tokens": bucket.nextRefillSize(),
			"last_refill":    bucket.lastRefill,
			"current_time":   time.Now(),
		}
	}
	return data
}

func (tbl *TBLimiter) GetLimit() int { return tbl.config.Capacity }
