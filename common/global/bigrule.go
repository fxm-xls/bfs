package global

import (
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/registry"
	"gorm.io/gorm"
)

const (
	ProjectName = "BigRule"
	// bigrule Version Info
	Version = "v1"
)

var GinEngine *gin.Engine
var CasbinEnforcer *casbin.SyncedEnforcer
var EtcdReg registry.Registry

var DBMysql *gorm.DB

var Logo = []byte{
	10, 32, 32, 32, 32, 47, 47, 32, 32, 32, 41, 32, 41, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32,
	32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 47, 47, 32, 32, 32, 41, 32, 41, 10, 32, 32, 32, 47, 47,
	95, 95, 95, 47, 32, 47, 32, 32, 32, 32, 32, 40, 32, 41, 32, 32, 32, 32, 32, 95, 95, 95, 32, 32, 32,
	32, 32, 32, 32, 47, 47, 95, 95, 95, 47, 32, 47, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32,
	32, 32, 32, 47, 47, 32, 32, 32, 32, 32, 32, 95, 95, 95, 10, 32, 32, 47, 32, 95, 95, 32, 32, 40, 32,
	32, 32, 32, 32, 32, 47, 32, 47, 32, 32, 32, 32, 47, 47, 32, 32, 32, 41, 32, 41, 32, 32, 32, 47, 32,
	95, 95, 95, 32, 40, 32, 32, 32, 32, 32, 32, 47, 47, 32, 32, 32, 47, 32, 47, 32, 32, 32, 47, 47, 32,
	32, 32, 32, 32, 47, 47, 95, 95, 95, 41, 32, 41, 10, 32, 47, 47, 32, 32, 32, 32, 41, 32, 41, 32, 32,
	32, 32, 47, 32, 47, 32, 32, 32, 32, 40, 40, 95, 95, 95, 47, 32, 47, 32, 32, 32, 47, 47, 32, 32, 32,
	124, 32, 124, 32, 32, 32, 32, 32, 47, 47, 32, 32, 32, 47, 32, 47, 32, 32, 32, 47, 47, 32, 32, 32,
	32, 32, 47, 47, 10, 47, 47, 95, 95, 95, 95, 47, 32, 47, 32, 32, 32, 32, 47, 32, 47, 32, 32, 32, 32,
	32, 32, 47, 47, 95, 95, 32, 32, 32, 32, 32, 47, 47, 32, 32, 32, 32, 124, 32, 124, 32, 32, 32, 32,
	40, 40, 95, 95, 95, 40, 32, 40, 32, 32, 32, 47, 47, 32, 32, 32, 32, 32, 40, 40, 95, 95, 95, 95, 10,
}
