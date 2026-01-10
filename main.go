package main

import (
	"context"
	"flag"
	"fmt"
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
	// Parse command-line flags
	migrateCmd := flag.NewFlagSet("migrate", flag.ExitOnError)
	migrateVersion := migrateCmd.String("version", "", "Migration version to run")
	migrateAction := migrateCmd.String("action", "up", "Migration action: up or down")

	// init
	config.Init()

	// Check if we're running a migration command
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		migrateCmd.Parse(os.Args[2:])
		if *migrateVersion == "" {
			log.Fatal(log.WorkerCtx, "Migration version is required. Use --version flag")
		}
		if err := runMigration(*migrateVersion, *migrateAction); err != nil {
			log.Fatalf(log.WorkerCtx, "Failed to run migration: %v", err)
		}
		log.Infof(log.WorkerCtx, "Migration %s %s completed successfully", *migrateVersion, *migrateAction)
		return
	}

	// Run all migrations (normal startup)
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

func runMigration(version string, action string) error {
	migrator := migration.NewMigrator(config.ContextDB(log.WorkerCtx))

	// Add all migrations
	migrations := migration.GetAllMigrations()
	for _, m := range migrations {
		migrator.AddMigration(m)
	}

	// Find the specific migration
	var targetMigration *migration.MigrationStep
	for _, m := range migrations {
		if m.Version == version {
			targetMigration = &m
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration version %s not found", version)
	}

	// Run the migration action
	switch action {
	case "up":
		log.Infof(log.WorkerCtx, "Running migration %s up: %s", version, targetMigration.Name)
		return migrator.RunOne(version, true)
	case "down":
		log.Infof(log.WorkerCtx, "Running migration %s down: %s", version, targetMigration.Name)
		return migrator.RunOne(version, false)
	default:
		return fmt.Errorf("invalid action: %s. Use 'up' or 'down'", action)
	}
}
