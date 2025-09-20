package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/log"
	"github.com/EricWvi/sonora/migration"
	"github.com/EricWvi/sonora/service"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	// init
	config.Init()

	// Run migrations
	if err := runMigrations(); err != nil {
		log.Fatalf(log.WorkerCtx, "Failed to run migrations: %v", err)
	}

	// Set up HTTP server
	g := gin.New()
	Load(g)
	addr := viper.GetString("addr")

	server := &http.Server{
		Addr:    addr,
		Handler: g,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Infof(log.WorkerCtx, "Start to listening the incoming requests on http address: %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorf(log.WorkerCtx, "HTTP server error: %v", err)
		}
	}()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	log.Info(log.WorkerCtx, "Received shutdown signal, starting graceful shutdown...")

	// Shutdown HTTP server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Errorf(log.WorkerCtx, "HTTP server shutdown error: %v", err)
	}

	// Wait for all workers to finish with timeout
	done := make(chan struct{})
	go func() {
		service.WorkerCancel()
		service.WorkerWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info(log.WorkerCtx, "All workers stopped gracefully")
	case <-time.After(30 * time.Second):
		log.Warn(log.WorkerCtx, "Workers shutdown timeout, forcing exit")
	}

	log.Info(log.WorkerCtx, "Application shutdown complete")
}

func runMigrations() error {
	migrator := migration.NewMigrator(config.ContextDB(log.WorkerCtx))

	// Add all migrations
	migrations := migration.GetAllMigrations()
	for _, m := range migrations {
		migrator.AddMigration(m)
	}

	// Run migrations
	if err := migrator.Up(); err != nil {
		return err
	}

	return nil
}
