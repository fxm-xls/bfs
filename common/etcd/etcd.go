package etcd

import (
	"bigrule/common/global"
	"fmt"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/spf13/viper"
)

type Etcd struct {
	Host     string
	Port     int
}

func InitEtcd(cfg *viper.Viper) *Etcd {
	return &Etcd{
		Host:           cfg.GetString("host"),
		Port:          cfg.GetInt("port"),

	}
}

var EtcdConfig = new(Etcd)


func Setup(){
	//etcd init
	addr := EtcdConfig.Host + ":" +fmt.Sprint(EtcdConfig.Port)
	global.EtcdReg = etcd.NewRegistry(
		registry.Addrs(addr),
	)
}
