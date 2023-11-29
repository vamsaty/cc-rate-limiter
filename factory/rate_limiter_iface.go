package factory

type RateConfig map[string]string

type RateLimiter interface {
	CanLimit(string) error
	GetLimit() int
	Unregister(string)
	Stop()
	Stats() interface{}
}

type RateLimiterAlgo uint

const (
	NoLimitAlgo RateLimiterAlgo = iota
	TokenBucket
	FixedWindowCounter
)

func NewRateLimiter(algo RateLimiterAlgo, config RateConfig) RateLimiter {
	switch algo {
	case TokenBucket:
		return NewTokenBucketLimiter(config)
	case FixedWindowCounter:
		return NewWindowLimiter(config)
	default:
		return &DummyRateLimit{}
	}
}

func NewRateLimiterFromConfig(config RateConfig) RateLimiter {
	switch config["algo"] {
	case "token_bucket":
		return NewTokenBucketLimiter(config)
	case "fixed_window_counter":
		return NewWindowLimiter(config)
	default:
		return &DummyRateLimit{}
	}
}

// DummyRateLimit is a dummy rate limiter that does nothing
type DummyRateLimit struct{}

func (d *DummyRateLimit) CanLimit(_ string) error { return nil }
func (d *DummyRateLimit) Unregister(_ string)     {}
func (d *DummyRateLimit) Stop()                   {}
func (d *DummyRateLimit) Stats() interface{}      { return nil }
func (d *DummyRateLimit) GetLimit() int           { return 1e9 }
