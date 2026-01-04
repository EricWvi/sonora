package config

import (
	"log/slog"
	"os"

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

	// If CONFIG_PATH is set, use it as the primary config path
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		viper.AddConfigPath(configPath)
	}

	// Default search paths (still used if CONFIG_PATH is not set or file not found there)
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
}
