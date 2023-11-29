package factory

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

// WindowConfig is the configuration for a window
type WindowConfig struct {
	// WindowSize is the size of the window (interval)
	WindowSize time.Duration
	// MaxRequestCount is the maximum number of requests allowed in a window
	MaxRequestCount int64
}

type WindowLimiterImpl struct {
	// windows is a map of windowId to window
	*sync.Mutex
	windowMap map[string]*window
	config    *WindowConfig
	shutdown  chan struct{} // is dummy - unused
}

func (w *WindowLimiterImpl) CanLimit(id string) error {
	w.Lock()
	defer w.Unlock()

	fwc := w.windowMap[id]
	if fwc == nil {
		fwc = &window{
			Id:              id,
			windowSize:      w.config.WindowSize,
			maxRequestCount: w.config.MaxRequestCount,
			requestCount:    0,
			startTime:       time.Now(),
		}
		w.windowMap[id] = fwc
	}
	return fwc.allowRequest()
}

func (w *WindowLimiterImpl) Unregister(s string) {
	w.Lock()
	defer w.Unlock()
	if fwc := w.windowMap[s]; fwc != nil {
		delete(w.windowMap, s)
	}
}

func (w *WindowLimiterImpl) Stop() {
	w.Lock()
	defer w.Unlock()
	close(w.shutdown)
}

func (w *WindowLimiterImpl) Stats() interface{} {
	data := make(map[string]interface{})
	for key, fwc := range w.windowMap {
		data[key] = map[string]interface{}{
			"map":      w.windowMap,
			"size":     fwc.requestCount,
			"capacity": fwc.maxRequestCount,
			"interval": fwc.windowSize,
		}
	}
	return data
}

func (w *WindowLimiterImpl) GetLimit() int {
	return int(w.config.MaxRequestCount)
}

func NewWindowLimiter(config map[string]string) *WindowLimiterImpl {
	var err error

	winConfig := &WindowConfig{}

	// parse WindowSize
	winConfig.WindowSize, err = time.ParseDuration(config["window_size"])
	if err != nil {
		panic(err)
	}

	// parse MaxRequestCount
	winConfig.MaxRequestCount, err = strconv.ParseInt(config["max_request_count"], 10, 64)
	if err != nil {
		panic(err)
	}

	fmt.Println("WINDOW", winConfig)
	return &WindowLimiterImpl{
		Mutex:     &sync.Mutex{},
		windowMap: make(map[string]*window),
		config:    winConfig,
		shutdown:  make(chan struct{}),
	}
}
