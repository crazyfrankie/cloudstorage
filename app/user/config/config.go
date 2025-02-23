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
	JWT    JWT    `yaml:"jwt"`
	ETCD   ETCD   `yaml:"etcd"`
	Minio  Minio  `yaml:"minio"`
}

type Server struct {
	Addr string `yaml:"addr"`
}

type MySQL struct {
	DSN string `yaml:"dsn"`
}

type JWT struct {
	Secret string `yaml:"secret"`
}

type ETCD struct {
	Addr string `yaml:"addr"`
}

type Minio struct {
	EndPoint    string `yaml:"endPoint"`
	AccessKey   string `yaml:"accessKey"`
	SecretKey   string `yaml:"secretKey"`
	BucketName  string `yaml:"bucketName"`
	DefaultName string `yaml:"defaultName"`
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
	if env == "" {
		return "test"
	}

	return env
}
