package factory

import (
	tokenBucket "githum.com/vamsaty/cc-rate-limiter/token-bucket"
)

type RateConfig map[string]string

type RateLimiter interface {
	CanLimit(string) error
	Unregister(string)
	Stop()
	Stats() interface{}
}

type RateLimiterAlgo uint

const (
	InvalidAlgo RateLimiterAlgo = iota
	TokenBucket
	LeakyBucketMeter
	LeakyBucketQueue
	SlidingWindowCounter
)

func NewRateLimiter(algo RateLimiterAlgo, config RateConfig) RateLimiter {
	switch algo {
	case TokenBucket:
		return tokenBucket.NewTokenBucketLimiter(config)
	default:
		return nil
	}
}

func NewRateLimiterFromConfig(config RateConfig) RateLimiter {
	switch config["algo"] {
	case "token_bucket":
		return tokenBucket.NewTokenBucketLimiter(config)
	default:
		return nil
	}
}
