package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/handler/album"
	"github.com/EricWvi/sonora/handler/media"
	"github.com/EricWvi/sonora/handler/singer"
	"github.com/EricWvi/sonora/handler/sync"
	"github.com/EricWvi/sonora/handler/track"
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

	// serve index.html at root path
	g.StaticFile("/", viper.GetString("route.front.index"))
	g.StaticFile("/admin/", viper.GetString("route.front.admin"))

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

	back.GET("/singer", singer.DefaultHandler)
	back.POST("/singer", singer.DefaultHandler)

	back.GET("/album", album.DefaultHandler)
	back.POST("/album", album.DefaultHandler)

	back.GET("/track", track.DefaultHandler)
	back.POST("/track", track.DefaultHandler)

	back.GET("/sync", sync.DefaultHandler)

	// Handle 404 for all unmatched routes
	g.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"code":    404,
			"message": "404 page not found - Gin Server",
		})
	})

	return g
}
