package config

import (
	"bigrule/common/etcd"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// database config
var cfgDatabase *viper.Viper

// application config
var cfgApplication *viper.Viper

// log config
var cfgLogger *viper.Viper

// etcd config
var cfgEtcd *viper.Viper

// user config
var cfgUser *viper.Viper

//setup config
func Setup(path string) {
	viper.SetConfigFile(path)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(fmt.Sprintf("Read config file fail: %s", err.Error()))
	}
	//Replace environment variables
	err = viper.ReadConfig(strings.NewReader(os.ExpandEnv(string(content))))
	if err != nil {
		log.Fatal(fmt.Sprintf("Parse config file fail: %s", err.Error()))
	}
	//DbMysql
	cfgDatabase = viper.Sub("bigrule.repo-bfs-service.dbmysql")
	if cfgDatabase == nil {
		panic("No found bigrule.repo-bfs-service.dbmysql in the configuration")
	}
	DbMysqlConfig = InitDbMysql(cfgDatabase)
	//application
	cfgApplication = viper.Sub("bigrule.repo-bfs-service.application")
	if cfgApplication == nil {
		panic("No found bigrule.repo-bfs-service.application in the configuration")
	}
	ApplicationConfig = InitApplication(cfgApplication)
	//logger
	cfgLogger = viper.Sub("bigrule.repo-bfs-service.logger")
	if cfgLogger == nil {
		panic("No found bigrule.repo-bfs-service.logger in the configuration")
	}
	LoggerConfig = InitLogger(cfgLogger)
	//etcd
	cfgEtcd = viper.Sub("bigrule.etcd-service")
	if cfgEtcd == nil {
		panic("No found bigrule.etcd-service in the configuration")
	}
	etcd.EtcdConfig = etcd.InitEtcd(cfgEtcd)
	//user
	cfgUser = viper.Sub("bigrule.repo-bfs-service.repo-user")
	if cfgUser == nil {
		panic("No found bigrule.repo-bfs-service.repo-user in the configuration")
	}
	UserConfig = InitUser(cfgUser)
	//......
}
