package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Server struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type Config struct {
	Server           Server `mapstructure:"server"`
	DockerComposeDir string `mapstructure:"docker-compose-directory"`
}

var config Config

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("can not get user home directory %v", err))
	}

	viper.AddConfigPath(home)
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	viper.SetConfigName("abshar-toolbox")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic("Not Exists abshar-toolbox.yaml")
	}

	cnf := new(Config)
	if err := viper.Unmarshal(cnf); err != nil {
		panic(fmt.Errorf("Unmarshal config Failed %v", err))
	}

	config = *cnf
}

func GetCnf() Config {
	return config
}

func Get(path string) string {
	return viper.GetString(path)
}
