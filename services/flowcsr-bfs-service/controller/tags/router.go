package tags

import (
	"bigrule/common/global"
	"bigrule/common/ico"
	"bigrule/services/flowcsr-bfs-service/middleware"
	"fmt"
	"github.com/gin-gonic/gin"
)

type TagRouter struct{}

func (sr TagRouter) Router(router *gin.Engine) {
	r := router.Group(fmt.Sprintf("/%s/tags", global.Version)).Use(middleware.AuthToken())
	{
		r.POST("/query", ico.Handler(TagQuery{}))
		r.POST("/delete", ico.Handler(TagDelete{}))
		r.POST("/export", ico.Handler(TagExport{}))
	}
}
