package main

import (
	"github.com/gin-gonic/gin"
	"githum.com/vamsaty/cc-rate-limiter/factory"
)

type Server struct {
	factory.RateLimiter
}

func NewServer(rateLimiter factory.RateLimiter) *Server {
	return &Server{RateLimiter: rateLimiter}
}

func Pack(code int, before, after interface{}) map[string]interface{} {
	return map[string]interface{}{
		"before": before,
		"after":  after,
		"status": code,
	}
}

func (s *Server) Start() error {
	r := gin.Default()
	r.GET("/limited", func(c *gin.Context) {
		before := s.RateLimiter.Stats()
		if s.RateLimiter.CanLimit(c.ClientIP()) != nil {
			c.IndentedJSON(429, Pack(429, before, s.RateLimiter.Stats()))
		} else {
			c.IndentedJSON(200, Pack(200, before, s.RateLimiter.Stats()))
		}
	})
	return r.Run(":8080")
}

func main() {
	limiter := factory.NewRateLimiter(
		factory.TokenBucket, map[string]string{
			"bucket_capacity":     "10",
			"token_push_interval": "5s",
		},
	)
	server := NewServer(limiter)
	server.Start()
}
