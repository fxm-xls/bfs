package ping

import (
	"bigrule/common/ico"
	"github.com/gin-gonic/gin"
)

type Ping struct {}

func (p Ping) DoHandle(c *gin.Context) *ico.Result {
	return ico.Succ("ping success to flowcsr-bfs-service...")
}
