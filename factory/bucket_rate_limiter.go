package factory

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

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
		// start with a filled bucket
		size.Store(tbl.config.BucketCapacity)
		tb := &tokenBucket{
			Id:                Id,
			bucketCapacity:    tbl.config.BucketCapacity,
			bucketSizeAtomic:  size,
			tokenPushInterval: tbl.config.TokenPushInterval,
			stopper:           make(chan struct{}),
		}
		tbl.bucketMap[Id] = tb
		// start the token pusher for this *new user*
		go tb.startTokenPusher()
		bucket = tb // update the bucket
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
			"map":      tbl.bucketMap,
			"size":     bucket.bucketSizeAtomic.Load(),
			"capacity": bucket.bucketCapacity,
			"interval": bucket.tokenPushInterval,
		}
	}
	return data
}

func (tbl *TBLimiter) GetLimit() int { return int(tbl.config.BucketCapacity) }

// NewTokenBucketLimiter creates a new token bucket limiter
func NewTokenBucketLimiter(config map[string]string) *TBLimiter {
	var err error
	tbConfig := &TokenBucketConfig{}

	// parse the token push interval
	tbConfig.TokenPushInterval, err = time.ParseDuration(config["token_push_interval"])
	if err != nil {
		panic(err)
	}
	fmt.Println("token push interval", tbConfig.TokenPushInterval)

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
