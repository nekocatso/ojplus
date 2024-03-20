package config

import (
	"github.com/spf13/viper"
)

type GlobalConfig struct {
	Mysql *MysqlConfig
}

type MysqlConfig struct {
	Host     string
	User     string
	Password string
}

func NewConfig(configPath, configName string) (*GlobalConfig, error) {
	var config GlobalConfig
	viper.SetConfigName(configName)
	viper.AddConfigPath(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	viper.Unmarshal(&config)
	return &config, nil
}
