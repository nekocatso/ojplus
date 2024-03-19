package config

import (
	"log"

	"github.com/spf13/viper"
)

// 读取配置文件config
type Config struct {
	Crypto   CryptoConfig
	Database DataBaseConfig
}

type CryptoConfig struct {
	RSAPublicKeyPath string
}
type DataBaseConfig struct {
	Address string
}

func NewConfig(path string) *Config {
	var config Config
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config file: %s \n", err)
	}

	viper.Unmarshal(&config)
	return &config
}
