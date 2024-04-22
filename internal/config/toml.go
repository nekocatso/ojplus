package config

import (
	"Ojplus/internal/email"

	"github.com/spf13/viper"
)

type Global struct {
	Gin *Gin
}

// Gin
type Gin struct {
	Port  string
	Mysql *Mysql
	Redis *Redis
	Token *Token
	Auth  *Auth
	Email *email.Email
}

type Mysql struct {
	DSN string
}

type Redis struct {
	Addr     string
	Password string
	DB       int
}
type Token struct {
	PrivateKeyPath string
	RefreshExp     int
	AccessExp      int
}

type Auth struct {
	VerificationExp int
}

func NewConfig(configPath, configName string) (*Global, error) {
	var config Global
	viper.SetConfigName(configName)
	viper.AddConfigPath(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	viper.Unmarshal(&config)
	return &config, nil
}
