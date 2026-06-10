// Command api starts the sandbox payment gateway HTTP server.
package main

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rmalik/sandbox-payments/internal/api"
	"github.com/rmalik/sandbox-payments/internal/config"
	"github.com/rmalik/sandbox-payments/internal/payment"
	"github.com/rmalik/sandbox-payments/internal/storage"
	"github.com/rmalik/sandbox-payments/internal/version"
	"github.com/rmalik/sandbox-payments/internal/webhook"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg := config.Load()

	if err := os.MkdirAll(filepath.Dir(cfg.DatabasePath), 0o755); err != nil {
		log.Error("create data directory", "error", err)
		os.Exit(1)
	}

	store, err := storage.OpenSQLite(cfg.DatabasePath)
	if err != nil {
		log.Error("open database", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	tmpl, err := template.ParseGlob(filepath.Join(cfg.TemplateDir, "*.html"))
	if err != nil {
		log.Error("parse templates", "error", err)
		os.Exit(1)
	}

	payments := payment.NewService(store, cfg.BaseURL)
	webhooks := webhook.NewSender(log)
	handler := api.NewHandler(payments, webhooks, tmpl, cfg.BaseURL, cfg.DocsDir, cfg.ChangelogPath, log)
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      api.NewServer(handler, cfg.StaticDir, log),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("server starting", "addr", server.Addr, "base_url", cfg.BaseURL, "version", version.Version)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error("shutdown failed", "error", err)
	}
	log.Info("server stopped")
}
