package router

import (
	"bigrule/common/global"
	"bigrule/common/ico"
	"github.com/gin-gonic/gin"
)

func init() {
	global.GinEngine = gin.New()
	global.GinEngine.Use(gin.Recovery())
	global.GinEngine.Use(rCros)
	global.GinEngine.NoRoute(rNoRoute)
}

func rCros(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "PUT, POST, GET, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Authorization, Accept, X-Requested-With, Token, Timestamp, Source, x-access-token")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(200)
		return
	}

	if c.Request.URL.Path == "/favicon.ico" {
		c.AbortWithStatus(200)
		return
	}
	c.Next()
}

func rNoRoute(c *gin.Context) {
	rst := ico.Err(404, "Page Not Found")
	c.AbortWithStatusJSON(200, rst)
}

type IMRouter interface {
	Router(c *gin.Engine)
}

func RouterRegister(routes ...IMRouter) {
	for _, r := range routes {
		r.Router(global.GinEngine)
	}
}