package config

import (
	"log"

	"github.com/spf13/viper"
	"github.com/yuanyu90221/airline-order-system/internal/util"
)

type Config struct {
	Port          string `mapstructure:"PORT"`
	RedisAddr     string `mapstructure:"REDIS_ADDR"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	GinMode       string `mapstructure:"GIN_MODE"`
	DbURL         string `mapstructure:"DB_URL"`
	DefaultTotal  int    `mapstructure:"DEFAULT_TOTAL"`
	DefaultWait   int    `mapstructure:"DEFAULT_WAIT"`
}

var AppConfig *Config

func init() {
	v := viper.New()
	v.AddConfigPath(".")
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()
	util.FailOnError(v.BindEnv("PORT"), "Failed on Bind PORT")
	util.FailOnError(v.BindEnv("REDIS_ADDR"), "Failed on Bind REDIS_ADDR")
	util.FailOnError(v.BindEnv("REDIS_PASSWORD"), "Failed on Bind REDIS_PASSWORD")
	util.FailOnError(v.BindEnv("GIN_MODE"), "Failed on Bind GIN_MODE")
	util.FailOnError(v.BindEnv("DB_URL"), "Failed on Bind DB_URL")
	util.FailOnError(v.BindEnv("DEFAULT_TOTAL", "DEFAULT_WAIT"), "Failed on bind DEFAULT_TOTAL, DEFAULT_WAIT")
	err := v.ReadInConfig()
	if err != nil {
		log.Println("Load from environment variable")
	}
	err = v.Unmarshal(&AppConfig)
	if err != nil {
		util.FailOnError(err, "Failed to read enivronment")
	}
}
