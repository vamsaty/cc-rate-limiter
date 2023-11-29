package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/vamsaty/cc-rate-limiter/factory"
	"sync"
)

type Server struct {
	limiterLock *sync.RWMutex
	r           *gin.Engine
	factory.RateLimiter
	PreviousRateLimiter factory.RateLimiter
}

func NewServer(rateLimiter factory.RateLimiter) *Server {
	return &Server{
		limiterLock: &sync.RWMutex{},
		RateLimiter: rateLimiter,
	}
}

func Pack(code int, before, after interface{}) map[string]interface{} {
	return map[string]interface{}{
		"before": before,
		"after":  after,
		"status": code,
	}
}

func (s *Server) Start() error {
	s.r = gin.Default()
	s.r.GET("/limited", func(c *gin.Context) {
		s.limiterLock.Lock()
		defer s.limiterLock.Unlock()

		before := s.RateLimiter.Stats()
		if s.RateLimiter.CanLimit(c.ClientIP()) != nil {
			c.IndentedJSON(429, Pack(429, before, s.RateLimiter.Stats()))
		} else {
			c.IndentedJSON(200, Pack(200, before, s.RateLimiter.Stats()))
		}
	})
	return s.r.Run(":8080")
}

func (s *Server) UpdateRateLimiter(limiter factory.RateLimiter) error {
	s.limiterLock.Lock()
	defer s.limiterLock.Unlock()
	s.PreviousRateLimiter = s.RateLimiter
	s.RateLimiter = limiter
	return nil
}

func (s *Server) Revert() {
	s.limiterLock.Lock()
	defer s.limiterLock.Unlock()
	s.RateLimiter = s.PreviousRateLimiter
}

var (
	rateLimitAlgo = flag.String("algo", "token_bucket", "rate limit algorithm")
	/*token bucket flags*/
	bucketCapacity    = flag.String("bucket_capacity", "100", "bucket capacity")
	tokenPushInterval = flag.String("token_push_interval", "5s", "token push interval")
)

func ParseRateLimiterConfig() factory.RateConfig {
	return factory.RateConfig{
		"algo":                *rateLimitAlgo,
		"bucket_capacity":     *bucketCapacity,
		"token_push_interval": *tokenPushInterval,
	}
}

func main() {
	flag.Parse()
	config := ParseRateLimiterConfig()
	limiter := factory.NewRateLimiterFromConfig(config)
	NewServer(limiter).Start()
}
