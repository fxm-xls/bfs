package repos

import (
	"bigrule/common/global"
	"bigrule/common/ico"
	"bigrule/services/flowcsr-bfs-service/middleware"
	"fmt"
	"github.com/gin-gonic/gin"
)

type RepoRouter struct{}

func (sr RepoRouter) Router(router *gin.Engine) {
	r := router.Group(fmt.Sprintf("/%s/repos", global.Version)).Use(middleware.AuthToken())
	{
		r.POST("/list/query", ico.Handler(RepoQuery{}))
		r.POST("/attributes/list/query", ico.Handler(RepoAttrQuery{}))
		r.POST("/dimensions/list/query", ico.Handler(RepoDimQuery{}))
	}
}
