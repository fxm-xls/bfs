package config

import "github.com/spf13/viper"

type DBmysql struct {
	Addr          string
	Loglevel	  string
}

func InitDbMysql(cfg *viper.Viper) *DBmysql {
	return &DBmysql{
		Addr:          cfg.GetString("addr"),
		Loglevel:      cfg.GetString("loglevel"),
	}
}

var DbMysqlConfig = new(DBmysql)