package factory

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrWindowFull = errors.New("window is full")
)

type window struct {
	// Id is used to identify the entity for which the window is created
	Id string
	// requestCount is the total number of requests in the window
	requestCount int64
	// maxRequestCount specifies the max number of requests allowed in the window
	maxRequestCount int64
	// windowSize specifies the window's length
	windowSize time.Duration
	// startTime specifies the start time of the window
	startTime time.Time
}

func (w *window) reset() {
	fmt.Println("resetting window for", w.Id)
	w.requestCount = 0
	w.startTime = time.Now()
}

func (w *window) allowRequest() error {
	// check window reset
	if time.Now().Sub(w.startTime) > w.windowSize {
		w.reset()
	}
	fmt.Println("W_COUNT", w.requestCount, "MAX", w.maxRequestCount)
	// check if window is full
	if w.requestCount >= w.maxRequestCount {
		return ErrWindowFull
	}
	w.requestCount++
	return nil
}
