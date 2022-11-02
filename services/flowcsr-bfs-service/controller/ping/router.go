package ping

import (
	"bigrule/common/ico"
	"github.com/gin-gonic/gin"
)

type PingRouter struct{}

func (sr PingRouter) Router(router *gin.Engine) {
	r := router.Group("/v1")
	{
		r.GET("/ping", ico.Handler(Ping{}))
	}
}
