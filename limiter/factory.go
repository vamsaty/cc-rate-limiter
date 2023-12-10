package limiter

import (
	ccUtils "github.com/vamsaty/cc-utils"
	"sync"
)

type RateConfig map[string]string

type RateLimiter interface {
	Allow(string) error
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

func NewRateLimiterFromConfig(config RateConfig) RateLimiter {
	switch config["algo"] {

	case "token_bucket":
		tbc := &TokenBucketConfig{}
		ccUtils.PanicIf(tbc.Parse(config))

		return &TBLimiter{
			bucketMap: make(map[string]*tokenBucket),
			RWMutex:   &sync.RWMutex{},
			shutDown:  make(chan struct{}),
			config:    tbc,
		}

	case "fixed_window_counter":
		winConfig := &WindowConfig{}
		ccUtils.PanicIf(winConfig.Parse(config))

		return &WindowLimiterImpl{
			Mutex:     &sync.Mutex{},
			windowMap: make(map[string]*window),
			config:    winConfig,
		}

	case "sliding_window_log":

		swlc := &SlidingWindowLogConfig{}
		if err := swlc.Parse(config); err != nil {
			panic(err)
		}
		return &SlidingWindowLogRateLimiter{
			config: swlc,
			swMap:  make(map[string]*slidingWindowLog),
		}
	default:
		return &DummyRateLimit{}
	}
}

// DummyRateLimit is a dummy rate limiter that does nothing
type DummyRateLimit struct{}

func (d *DummyRateLimit) Allow(_ string) error { return nil }
func (d *DummyRateLimit) Unregister(_ string)  {}
func (d *DummyRateLimit) Stop()                {}
func (d *DummyRateLimit) Stats() interface{}   { return nil }
func (d *DummyRateLimit) GetLimit() int        { return 1e9 }
