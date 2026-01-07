// Copyright (c) Kyle Huggins and contributors
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	goversion "github.com/caarlos0/go-version"

	"github.com/hugginsio/skopos/internal/config"
	"github.com/hugginsio/skopos/internal/tailscale"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(5)
	}

	handlerOptions := &slog.HandlerOptions{Level: slog.LevelInfo}
	if cfg.Development {
		handlerOptions.Level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stdout, handlerOptions)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	logger.Info("skopos dnsmon starting", "version", goversion.GetVersionInfo().GitVersion, "development", cfg.Development)

	// Create context that listens for termination signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handler for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	ts, err := tailscale.New(tailscale.Config{
		Hostname:  cfg.Tailscale.Hostname,
		StateDir:  cfg.Tailscale.StateDir,
		AuthKey:   cfg.Tailscale.AuthKey,
		Logger:    logger,
		Ephemeral: false,
	})

	if err != nil {
		logger.Error("failed to create Tailscale server", "error", err)
		os.Exit(5)
	}

	if err := ts.Start(ctx); err != nil {
		logger.Error("failed to start Tailscale", "error", err)
		os.Exit(5)
	}

	// TODO: Initialize RPC server
	// TODO: Initialize health checker
	// TODO: Initialize DNS config syncer

	logger.Info("dnsmon OK")

	// Wait for termination signal
	sig := <-sigChan
	logger.Info("received shutdown signal", "signal", sig)

	// TODO: Stop health checker
	// TODO: Stop DNS config syncer
	// TODO: Stop RPC server

	if err := ts.Close(); err != nil {
		logger.Error("error closing Tailscale server", "error", err)
	}

	logger.Info("goodbye")
}
