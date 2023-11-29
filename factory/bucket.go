package factory

import (
	"fmt"
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

// startTokenPusher starts the token pusher, pushes tokens into the bucket
// at a fixed interval specified by tokenPushInterval
func (tb *tokenBucket) startTokenPusher() {
	fmt.Println("starting token pusher for", tb.Id)
	ticker := time.NewTicker(tb.tokenPushInterval)
	for {
		select {
		case <-ticker.C:
			if tb.bucketSizeAtomic.Load() == tb.bucketCapacity {
				fmt.Println("bucket is full, not adding token")
				continue
			}
		case <-tb.stopper:
			fmt.Println("stopping token pusher for", tb.Id)
			return
		}
	}
}

// allowRequest checks if a request can be allowed
func (tb *tokenBucket) allowRequest() error {
	if tb.bucketSizeAtomic.Load() <= 0 {
		return ErrBucketEmpty
	}
	// remove a token
	tb.bucketSizeAtomic.Add(-1)
	return nil
}
