package main

import (
	"fmt"
	"github.com/vamsaty/cc-rate-limiter/factory"
	"net/http"
	"net/url"
	"testing"
	"time"
)

var getStatusCodeMap = func(numReq int) map[int]int {
	client := http.DefaultClient
	respMap := map[int]int{}

	for i := 0; i < numReq; i++ {
		resp, _ := client.Do(&http.Request{
			Method: http.MethodGet,
			URL:    &url.URL{Scheme: "http", Host: "localhost:8080", Path: "/limited"},
		})
		respMap[resp.StatusCode]++
	}
	return respMap
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func IsValid(limit, numReq int) error {
	c200 := Min(numReq, limit)
	c429 := Max(0, numReq-limit)
	data := getStatusCodeMap(numReq)
	if data[200] == c200 && data[429] == c429 {
		return nil
	}
	return fmt.Errorf("expected 200=%d and 429=%d, got 200=%d and 429=%d", c200, c429, data[200], data[429])
}

type TestCase struct {
	name   string
	config factory.RateConfig
	numReq int
}

func updateAndExecuteTests(t *testing.T, testCases []TestCase) {
	server := NewServer(&factory.DummyRateLimit{})
	go server.Start()

	time.Sleep(1 * time.Second)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// NOTE: this supports updating the RateLimiter for the server on the fly
			// only for easier testing.
			limiter := factory.NewRateLimiterFromConfig(tc.config)
			_ = server.UpdateRateLimiter(limiter)

			if err := IsValid(limiter.GetLimit(), tc.numReq); err != nil {
				t.Fatal(err)
			}

		})
	}
}

func TestFixedWindowCounter(t *testing.T) {
	updateAndExecuteTests(
		t,
		[]TestCase{
			/* ---------- FixedWindowCount ---------- */
			{
				name: "[FixedWindowCount] Block requests",
				config: factory.RateConfig{
					"algo":              "fixed_window_counter",
					"max_request_count": "0",
					"window_size":       "10s",
				},
				numReq: 5,
			},
			{
				name: "[FixedWindowCount] Large window size, num requests more than window size",
				config: factory.RateConfig{
					"algo":              "fixed_window_counter",
					"max_request_count": "5",
					"window_size":       "10s",
				},
				numReq: 15,
			},
			{
				name: "[FixedWindowCount] Large window size, num requests less than window size",
				config: factory.RateConfig{
					"algo":                "token_bucket",
					"bucket_capacity":     "10",
					"token_push_interval": "10s",
				},
				numReq: 9,
			},
		},
	)
}

func TestTokenBucket(t *testing.T) {
	updateAndExecuteTests(
		t,
		[]TestCase{
			/* ---------- TokenBucket ---------- */
			{
				name: "[TokenBucket] Equal bucket size and client requests, with slow token filling",
				config: factory.RateConfig{
					"algo":                "token_bucket",
					"bucket_capacity":     "5",
					"token_push_interval": "10s",
				},
				numReq: 5,
			},
			{
				name: "[TokenBucket] Smaller bucket size and client requests, with slow token filling",
				config: factory.RateConfig{
					"algo":                "token_bucket",
					"bucket_capacity":     "9",
					"token_push_interval": "10s",
				},
				numReq: 10,
			},
			{
				name: "[TokenBucket] Block requests",
				config: factory.RateConfig{
					"algo":                "token_bucket",
					"bucket_capacity":     "0",
					"token_push_interval": "10s",
				},
				numReq: 1,
			},
		},
	)
}

func TestSlidingWindowLog(t *testing.T) {

}
