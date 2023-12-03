package limiter

import (
	"github.com/gin-gonic/gin"
	"sync"
)

type Server struct {
	limiterLock *sync.RWMutex
	r           *gin.Engine
	RateLimiter
	PreviousRateLimiter RateLimiter
}

func NewServer(rateLimiter RateLimiter) *Server {
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

func (s *Server) Start(address string) error {
	router := gin.New()
	router.Use(
		gin.LoggerWithWriter(gin.DefaultWriter, "/limited"),
		gin.Recovery(),
	)
	s.r = router
	s.r.GET("/limited", func(c *gin.Context) {
		s.limiterLock.Lock()
		defer s.limiterLock.Unlock()
		before := s.RateLimiter.Stats()
		id := c.GetHeader("X-User")
		if s.RateLimiter.Allow(id) != nil {
			c.IndentedJSON(429, Pack(429, before, s.RateLimiter.Stats()))
		} else {
			c.IndentedJSON(200, Pack(200, before, s.RateLimiter.Stats()))
		}
	})
	s.r.GET("/unlimited", func(c *gin.Context) {
		c.IndentedJSON(200, Pack(200, nil, nil))
	})
	s.r.GET("/stats", func(c *gin.Context) {
		c.IndentedJSON(200, s.RateLimiter.Stats())
	})
	return s.r.Run(address)
}

func (s *Server) UpdateRateLimiter(limiter RateLimiter) error {
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

func ParseRateLimiterConfig(args ...string) RateConfig {
	return RateConfig{
		"algo":              args[0],
		"capacity":          args[1],
		"refill_rate":       args[2],
		"max_request_count": args[3],
		"window_size":       args[4],
		"request_per_sec":   args[5],
	}
}
