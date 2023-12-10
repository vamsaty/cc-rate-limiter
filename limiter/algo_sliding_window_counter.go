package limiter

import (
	"fmt"
	"time"
)

// TODO: implement sliding window counter algorithm

var (
	ErrTooManyRequests = fmt.Errorf("too many requests")
)

type slidingWindowCounter struct {
	maxRequestPerWindow int
	// windowLen is size of the sliding window
	windowLen     time.Duration
	previousCount int
	currentCount  int
}

func (swc *slidingWindowCounter) allowRequest() error {
	if swc.currentCount >= swc.maxRequestPerWindow {
		return ErrTooManyRequests
	}
	return nil
}
