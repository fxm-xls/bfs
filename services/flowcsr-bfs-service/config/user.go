package config

import "github.com/spf13/viper"

type User struct {
	Name string
	Pwd  string
}

func InitUser(cfg *viper.Viper) *User {
	return &User{
		Name: cfg.GetString("name"),
		Pwd:  cfg.GetString("pwd"),
	}
}

var UserConfig = new(User)
