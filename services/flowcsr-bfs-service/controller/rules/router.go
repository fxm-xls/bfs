package rules

import (
	"bigrule/common/global"
	"bigrule/common/ico"
	"bigrule/services/flowcsr-bfs-service/middleware"
	"fmt"
	"github.com/gin-gonic/gin"
)

type RuleRouter struct{}

func (sr RuleRouter) Router(router *gin.Engine) {
	r := router.Group(fmt.Sprintf("/%s/rules", global.Version)).Use(middleware.AuthToken())
	{
		r.POST("/add-batch", ico.Handler(RuleAdd{}))
		r.POST("/query", ico.Handler(RuleQuery{}))
		r.POST("/attributes/query", ico.Handler(RuleAttrQuery{}))
		r.POST("/regex/query", ico.Handler(RuleRegexQuery{}))
		r.POST("/delete", ico.Handler(RuleDelete{}))
	}
}
