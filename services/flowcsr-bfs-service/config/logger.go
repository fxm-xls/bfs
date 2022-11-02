package config

import "github.com/spf13/viper"

type Logger struct {
	Path          string
	Level          string
	Stdout          bool
}

func InitLogger(cfg *viper.Viper) *Logger {
	return &Logger{
		Path:           cfg.GetString("path"),
		Level:          cfg.GetString("level"),
		Stdout:         cfg.GetBool("stdout"),
	}
}

var LoggerConfig = new(Logger)
