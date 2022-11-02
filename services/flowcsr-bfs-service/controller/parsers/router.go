package parsers

import (
	"bigrule/common/global"
	"bigrule/common/ico"
	"bigrule/services/flowcsr-bfs-service/middleware"
	"fmt"
	"github.com/gin-gonic/gin"
)

type ParserRouter struct{}

func (sr ParserRouter) Router(router *gin.Engine) {
	r := router.Group(fmt.Sprintf("/%s/parsers", global.Version)).Use(middleware.AuthToken())
	{
		r.POST("/delete", ico.Handler(ParserDelete{}))
	}
}
