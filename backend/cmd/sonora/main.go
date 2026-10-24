// Command sonora serves the music library, its admin API, the admin
// console and the player from a single process.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sonora"
	"sonora/internal/admin"
	"sonora/internal/config"
	"sonora/internal/devmode"
	"sonora/internal/server"
	"sonora/internal/store"
)

// Set at build time with -ldflags "-X main.version=... -X main.commit=...".
var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to the YAML config file")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("sonora %s (%s)\n", version, commit)
		return
	}
	if err := run(*configPath); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	st, err := store.Open(cfg.DataDir, cfg.MediaDir)
	if err != nil {
		return err
	}
	defer st.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.Dev {
		if err := seedDevAdmin(ctx, st); err != nil {
			return err
		}
		slog.Warn("development mode: setup token skipped; fixed accounts enabled",
			"admin", devmode.AdminUsername, "player_user", devmode.Username)
	}
	authn, err := admin.New(ctx, st, admin.Config{PublicURL: cfg.PublicURL, Dev: cfg.Dev})
	if err != nil {
		return err
	}
	if token := authn.SetupToken(); token != "" {
		printSetupBanner(os.Stdout, cfg.PublicURL, token)
	}

	handler, err := server.New(st, authn, sonora.Player(), sonora.Admin(), server.Options{Version: version})
	if err != nil {
		return err
	}
	srv := &http.Server{
		Addr:              cfg.Listen,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errc := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", cfg.Listen, "data_dir", cfg.DataDir, "media_dir", cfg.MediaDir, "version", version)
		errc <- srv.ListenAndServe()
	}()

	select {
	case err := <-errc:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

// seedDevAdmin creates the fixed development admin on a fresh data dir.
func seedDevAdmin(ctx context.Context, st *store.Store) error {
	has, err := st.HasAdmin(ctx)
	if err != nil || has {
		return err
	}
	hash, err := admin.HashPassword(devmode.AdminPassword)
	if err != nil {
		return err
	}
	_, err = st.CreateFirstAdmin(ctx, devmode.AdminUsername, hash)
	return err
}

// printSetupBanner writes the setup token to standard output, i.e. the
// container log. It is the only credential Sonora ever logs.
func printSetupBanner(w io.Writer, publicURL, token string) {
	fmt.Fprintf(w, `
==================== Sonora 初始化 ====================
尚未创建管理员。在浏览器打开 %s/admin/
输入下面的初始化 token、用户名和密码创建管理员：

    %s

token 只在本次运行中有效，服务重启后会重新生成。
=======================================================

`, publicURL, token)
}
