package healthcheck

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthCheck struct {
	Engine *gin.Engine
}

func NewHealthcheck() *HealthCheck {
	gin.SetMode("release")
	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return &HealthCheck{
		Engine: r,
	}
}
