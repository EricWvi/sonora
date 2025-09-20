package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/handler/media"
	"github.com/EricWvi/sonora/log"
	"github.com/EricWvi/sonora/middleware"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// Load loads the middlewares, routes, handlers.
func Load(g *gin.Engine, mw ...gin.HandlerFunc) *gin.Engine {
	middleware.InitJWTMap()

	// Basic Middlewares.
	g.Use(gin.Recovery())
	g.Use(mw...)
	g.Use(gzip.Gzip(gzip.DefaultCompression))

	// serve front dist
	dir := viper.GetString("route.front.dir")
	err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				p := strings.TrimPrefix(path, dir)
				g.StaticFile(p, path)
			}
			return nil
		})
	if err != nil {
		log.Error(log.WorkerCtx, err.Error())
		os.Exit(1)
	}
	dir = viper.GetString("route.journal.dir")
	err = filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				p := strings.TrimPrefix(path, dir)
				g.StaticFile("/journal"+p, path)
			}
			return nil
		})
	if err != nil {
		log.Error(log.WorkerCtx, err.Error())
		os.Exit(1)
	}

	// serve index.html at root path
	g.StaticFile("/", viper.GetString("route.front.index"))
	g.StaticFile("/journal/", viper.GetString("route.journal.index"))

	g.GET("/ping", handler.Ping)
	// middleware.BodyWriter retrieves response body
	g.Use(middleware.BodyWriter())
	// middleware.JWT inject user ID
	g.Use(middleware.JWT())
	// middleware.Idempotency handles idempotency key
	g.Use(middleware.Idempotency())

	raw := g.Group(viper.GetString("route.back.base"))
	raw.POST("/upload", media.Upload)
	raw.GET("/m/:link", media.Serve)

	back := g.Group(viper.GetString("route.back.base"))
	// middleware.Logging logs request and response
	back.Use(middleware.Logging())

	// Handle 404 for all unmatched routes
	g.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"code":    404,
			"message": "404 page not found - Gin Server",
		})
	})

	return g
}
