package config

import (
	"log/slog"

	"github.com/EricWvi/sonora/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func Init() {
	// init config
	if err := LoadCfg(); err != nil {
		panic(err)
	}

	// logger
	if gin.Mode() == gin.ReleaseMode {
		log.InitLogger(slog.LevelInfo)
	} else {
		log.InitLogger(slog.LevelDebug)
	}

	InitDB()
}

func LoadCfg() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
}
