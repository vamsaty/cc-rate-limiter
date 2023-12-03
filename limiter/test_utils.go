package limiter

import (
	"fmt"
	ccUtils "github.com/vamsaty/cc-utils"
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

// IsValid checks the expected number of 200 and 429 status code responses
func IsValid(maxAllowedReq, numReq int) error {
	// 200 status code count should be minimum of numReq and maximum allowed
	c200 := ccUtils.Min(numReq, maxAllowedReq)
	// 429 status code count is 0 if numReq < maximum allowed requests. Otherwise,
	// its numReq - maxAllowedReq. i.e. the number of requests that were blocked.
	c429 := ccUtils.Max(0, numReq-maxAllowedReq)
	data := getStatusCodeMap(numReq)
	if data[200] == c200 && data[429] == c429 {
		return nil
	}
	return fmt.Errorf("expected 200=%d and 429=%d, got 200=%d and 429=%d", c200, c429, data[200], data[429])
}

type TestCase struct {
	// name is the name of the testcase
	name string
	// config of the rate limiter
	config RateConfig
	// numReq is the number of requests to make in a testcase
	numReq int
}

// runTestCases updates the rate limiter for the server and runs the test cases
func runTestCases(t *testing.T, testCases []TestCase, validatorFunc func(int, int) error) {
	server := NewServer(&DummyRateLimit{})
	go server.Start(":8080")

	// for safety - sleep for server to come up.
	time.Sleep(1 * time.Second)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// NOTE: this supports updating the RateLimiter for the server on the fly
			// only for easier testing.
			limiter := NewRateLimiterFromConfig(tc.config)
			_ = server.UpdateRateLimiter(limiter)

			if err := validatorFunc(limiter.GetLimit(), tc.numReq); err != nil {
				t.Fatal(err)
			}
		})
	}
}
