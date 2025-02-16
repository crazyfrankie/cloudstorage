package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
)

var (
	once   sync.Once
	config *Config
)

type Config struct {
	Env    string
	Server Server `yaml:"server"`
	MySQL  MySQL  `yaml:"mysql"`
	Redis  Redis  `yaml:"redis"`
	JWT    JWT    `yaml:"jwt"`
	Minio  Minio  `yaml:"minio"`
	ETCD   ETCD   `yaml:"etcd"`
}

type Server struct {
	Addr string `yaml:"addr"`
}

type MySQL struct {
	DSN string `yaml:"dsn"`
}

type Redis struct {
	Addr string `yaml:"addr"`
}

type JWT struct {
	Secret string `yaml:"secret"`
}

type ETCD struct {
	Addr string `yaml:"addr"`
}

type Minio struct {
	EndPoint  string `yaml:"endPoint"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
}

func GetConf() *Config {
	once.Do(initConfig)

	return config
}

func initConfig() {
	prefix := "config"
	filePath := filepath.Join(prefix, filepath.Join(getEnv(), "config.yaml"))
	viper.SetConfigFile(filePath)

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	config = new(Config)
	if err := viper.Unmarshal(&config); err != nil {
		panic(err)
	}

	config.Env = getEnv()
	fmt.Printf("%#v", config)
}

func getEnv() string {
	env := os.Getenv("GO_ENV")
	if env != "" {
		return "test"
	}

	return env
}
