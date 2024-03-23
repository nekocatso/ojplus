package config

import (
	"github.com/spf13/viper"
)

type GinConfig struct {
	Mysql *MysqlConfig
	Redis *RedisConfig
	Port  string
}

type MysqlConfig struct {
	DSN string
}
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func NewConfig(configPath, configName string) (*GinConfig, error) {
	var config GinConfig
	viper.SetConfigName(configName)
	viper.AddConfigPath(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	viper.Unmarshal(&config)
	return &config, nil
}
