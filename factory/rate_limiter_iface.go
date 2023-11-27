package factory

import (
	tokenBucket "github.com/vamsaty/cc-rate-limiter/token-bucket"
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
	NoLimitAlgo RateLimiterAlgo = iota
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
		return &DummyRateLimit{}
	}
}

func NewRateLimiterFromConfig(config RateConfig) RateLimiter {
	switch config["algo"] {
	case "token_bucket":
		return tokenBucket.NewTokenBucketLimiter(config)
	default:
		return &DummyRateLimit{}
	}
}

type DummyRateLimit struct{}

func (d *DummyRateLimit) CanLimit(s string) error { return nil }
func (d *DummyRateLimit) Unregister(s string)     {}
func (d *DummyRateLimit) Stop()                   {}
func (d *DummyRateLimit) Stats() interface{}      { return nil }
