package config

import "github.com/spf13/viper"

type Application struct {
	Host          string
	Port          string
	Mode          string
}

func InitApplication(cfg *viper.Viper) *Application {
	return &Application{
		Host:          cfg.GetString("host"),
		Port:          cfg.GetString("port"),
		Mode:          cfg.GetString("mode"),
	}
}

var ApplicationConfig = new(Application)