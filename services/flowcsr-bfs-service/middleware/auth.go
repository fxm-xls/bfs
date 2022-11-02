package middleware

import (
	"bigrule/services/flowcsr-bfs-service/middleware/queue"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"sync"
)

var opMutex sync.Mutex

func AuthToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Access-Token") //cookie中拿到token
		if token == "" || token == "undefined" {
			c.JSON(http.StatusOK, gin.H{
				"status":  0,
				"code":    302,
				"message": "cookies不存在 请重新登陆",
				"data":    "token: " + token,
			})
			c.Abort()
			return
		}
		// 增删改操作需要入队
		url := c.Request.RequestURI
		if !strings.HasSuffix(url, "delete") && !strings.HasSuffix(url, "add-batch") {
			c.Next()
			return
		}
		// 增删改操作入队，加锁
		item := queue.Item{C: c.Copy()}
		queue.Q().Enqueue(item)
		opMutex.Lock()
		item = queue.Q().Dequeue()
		item.C.Next()
		opMutex.Unlock()
	}
}
