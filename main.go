package main

import (
	"flag"
	"github.com/vamsaty/cc-rate-limiter/limiter"
)

var (
	rateLimitAlgo = flag.String("algo", "token_bucket", "rate limit algorithm")

	/*token bucket flags*/
	capacity   = flag.String("capacity", "10", "bucket capacity")
	refillRate = flag.String("refill_rate", "1", "tokens to push per second")

	/*fixed window counter flags*/
	maxRequestCount = flag.String("max_request_count", "10", "maximum number of requests per window size")

	/*sliding window log flags*/
	requestPerSec = flag.String("request_per_sec", "10", "maximum number of requests per second")

	// windowSize is the size of the window - used for "fixed window counter" and "sliding window log"
	windowSize = flag.String("window_size", "1s", "window size to capture the requests")
)

func main() {
	flag.Parse()
	config := limiter.ParseRateLimiterConfig(
		*rateLimitAlgo,
		*capacity,
		*refillRate,
		*maxRequestCount,
		*windowSize,
		*requestPerSec,
	)
	rl := limiter.NewRateLimiterFromConfig(config)
	limiter.NewServer(rl).Start(":8080")
}
