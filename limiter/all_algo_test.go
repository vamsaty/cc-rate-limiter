package limiter

/*
Basic sanity tests for all the rate limiting algorithms
*/

import (
	"testing"
)

func TestTokenBucket(t *testing.T) {
	runTestCases(
		t,
		[]TestCase{
			{
				name: "[TokenBucket] Equal bucket size and client requests, with slow token filling",
				config: RateConfig{
					"algo":        "token_bucket",
					"capacity":    "5",
					"refill_rate": "10",
				},
				numReq: 5,
			},
			{
				name: "[TokenBucket] Smaller bucket size and client requests, with slow token filling",
				config: RateConfig{
					"algo":        "token_bucket",
					"capacity":    "9",
					"refill_rate": "10",
				},
				numReq: 10,
			},
			{
				name: "[TokenBucket] Block requests",
				config: RateConfig{
					"algo":        "token_bucket",
					"capacity":    "0",
					"refill_rate": "10",
				},
				numReq: 1,
			},
		},
		IsValid,
	)
}

func TestNewSlidingWindowLogRateLimiter(t *testing.T) {
	runTestCases(t, []TestCase{
		{
			name: "Block requests",
			config: RateConfig{
				"algo":            "sliding_window_log",
				"window_size":     "10s",
				"request_per_sec": "2",
			},
			numReq: 100,
		},
	}, IsValid)
}

func TestFixedWindowCounter(t *testing.T) {
	runTestCases(
		t,
		[]TestCase{
			{
				name: "[FixedWindowCount] Block requests",
				config: RateConfig{
					"algo":              "fixed_window_counter",
					"max_request_count": "0",
					"window_size":       "10s",
				},
				numReq: 5,
			},
			{
				name: "[FixedWindowCount] Large window size, num requests more than window size",
				config: RateConfig{
					"algo":              "fixed_window_counter",
					"max_request_count": "5",
					"window_size":       "10s",
				},
				numReq: 15,
			},
		},
		IsValid,
	)
}
