package token_bucket

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrBucketEmpty = fmt.Errorf("bucket is empty")
	ErrBucketFull  = fmt.Errorf("bucket is full")
)

// TokenBucketConfig is the configuration for the token bucket
type TokenBucketConfig struct {
	BucketCapacity    int64
	TokenPushInterval time.Duration
}

// tokenBucket is a token bucket implementation for rate limiting
type tokenBucket struct {
	bucketSizeAtomic  atomic.Int64  // current number of tokens in the bucket
	bucketCapacity    int64         // max number of tokens in the bucket
	Id                string        // Id of the bucket - username/userid or IP address
	tokenPushInterval time.Duration // time interval to push tokens into the bucket
	stopper           chan struct{} // stop the token pusher
}

// Stop stops the token pusher
func (tb *tokenBucket) Stop() {
	close(tb.stopper)
}

// addToken adds a token to the bucket
func (tb *tokenBucket) addToken() error {
	if tb.bucketSizeAtomic.Load() <= tb.bucketCapacity {
		tb.bucketSizeAtomic.Add(1)
		return nil
	}
	return ErrBucketFull
}

// removeToken removes a token from the bucket
// this happens when a request is allowed
func (tb *tokenBucket) removeToken() error {
	if tb.bucketSizeAtomic.Load() > 0 {
		tb.bucketSizeAtomic.Add(-1)
		return nil
	}
	return ErrBucketEmpty
}

// StartTokenPusher starts the token pusher, pushes tokens into the bucket
// at a fixed interval specified by tokenPushInterval
func (tb *tokenBucket) StartTokenPusher() {
	fmt.Println("starting token pusher for", tb.Id)
	ticker := time.NewTicker(tb.tokenPushInterval)
	for {
		select {
		case <-ticker.C:
			if err := tb.addToken(); err != nil {
				if err == ErrBucketFull {
					fmt.Println("bucket is full, not adding token")
				}
			}
		case <-tb.stopper:
			fmt.Println("stopping token pusher for", tb.Id)
			return
		}
	}
}

// TBLimiter is a token bucket limiter, satisfying the RateLimiter interface
type TBLimiter struct {
	// bucketMap is a map of userId to token bucket. Generally a bucket is
	// created for each user/IP address
	bucketMap map[string]*tokenBucket
	// Mutex is a lock for bucketMap
	*sync.Mutex
	// shutDown is a channel to stop the token pusher
	shutDown chan struct{}
	// config is the configuration for the token bucket
	config *TokenBucketConfig
}

// CanLimit checks if a request can be allowed.
func (tbl *TBLimiter) CanLimit(Id string) error {
	fmt.Println("checking if request can be allowed for", Id)
	tbl.Lock()
	defer tbl.Unlock()

	bucket, ok := tbl.bucketMap[Id]
	if !ok {
		fmt.Println("creating new bucket for", Id)
		size := atomic.Int64{}
		size.Store(0)
		tb := &tokenBucket{
			Id:                Id,
			bucketCapacity:    tbl.config.BucketCapacity,
			bucketSizeAtomic:  size,
			tokenPushInterval: tbl.config.TokenPushInterval,
			stopper:           make(chan struct{}),
		}
		tbl.bucketMap[Id] = tb
		// start the token pusher for this *new user*
		go tb.StartTokenPusher()
		bucket = tb // update the bucket
	}
	return bucket.removeToken()
}

// Unregister remove a bucket from the @bucketMap
func (tbl *TBLimiter) Unregister(Id string) {
	tbl.Lock()
	defer tbl.Unlock()

	if bucket, ok := tbl.bucketMap[Id]; ok { // bucket exists
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
			"map":      tbl.bucketMap,
			"size":     bucket.bucketSizeAtomic.Load(),
			"capacity": bucket.bucketCapacity,
			"interval": bucket.tokenPushInterval,
		}
	}
	return data
}

// NewTokenBucketLimiter creates a new token bucket limiter
func NewTokenBucketLimiter(config map[string]string) *TBLimiter {
	var err error
	tbConfig := &TokenBucketConfig{}

	// parse the token push interval
	tbConfig.TokenPushInterval, err = time.ParseDuration(config["token_push_interval"])
	if err != nil {
		panic(err)
	}

	// parse the bucket capacity
	tbConfig.BucketCapacity, err = strconv.ParseInt(config["bucket_capacity"], 10, 64)
	if err != nil {
		panic(err)
	}

	return &TBLimiter{
		bucketMap: make(map[string]*tokenBucket),
		Mutex:     &sync.Mutex{},
		shutDown:  make(chan struct{}),
		config:    tbConfig,
	}
}
