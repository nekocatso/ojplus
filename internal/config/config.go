package config

import (
	"github.com/spf13/viper"
)

// 读取配置文件config
type Config struct {
	Mysql MysqlTable
}

type MysqlTable struct {
	Host     string
	User     string
	Password string
}

func NewConfig(configPath, configName string) (*Config, error) {
	var config Config
	viper.SetConfigName(configName)
	viper.AddConfigPath(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	viper.Unmarshal(&config)
	return &config, nil
}
