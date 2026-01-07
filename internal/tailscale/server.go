// Copyright (c) Kyle Huggins and contributors
// SPDX-License-Identifier: BSD-3-Clause

package tailscale

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"

	"tailscale.com/client/local"
	"tailscale.com/ipn"
	"tailscale.com/ipn/ipnstate"
	"tailscale.com/tsnet"
)

// Server wraps a tsnet.Server and provides lifecycle management and access
// to Tailscale APIs.
type Server struct {
	server        *tsnet.Server
	localClient   *local.Client
	hostname      string
	stateDir      string
	logger        *slog.Logger
	started       bool
	tailscaleIP   string
	tailscaleIPv6 string
}

// Config holds configuration for creating a new Server.
type Config struct {
	Hostname  string       // Hostname is the hostname to use for this Tailscale node.
	StateDir  string       // StateDir is the directory where Tailscale state is stored.
	AuthKey   string       // AuthKey is the Tailscale authentication key.
	Logger    *slog.Logger // Logger is the logger to use for tsnet operations.
	Ephemeral bool         // Ephemeral indicates whether this node should be ephemeral.
}

// New creates a new tsnet Server with the given configuration.
func New(cfg Config) (*Server, error) {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	// Ensure state directory exists
	if err := os.MkdirAll(cfg.StateDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	cfg.Logger.Debug("initializing Tailscale", "hostname", cfg.Hostname, "state_dir", cfg.StateDir, "ephemeral", cfg.Ephemeral)

	// Create tsnet server
	srv := &tsnet.Server{
		Hostname:  cfg.Hostname,
		Dir:       cfg.StateDir,
		AuthKey:   cfg.AuthKey,
		Ephemeral: cfg.Ephemeral,
		Logf: func(format string, args ...any) {
			cfg.Logger.Debug(fmt.Sprintf(format, args...))
		},
	}

	return &Server{
		server:      srv,
		localClient: &local.Client{},
		hostname:    cfg.Hostname,
		stateDir:    cfg.StateDir,
		logger:      cfg.Logger,
		started:     false,
	}, nil
}

// Start starts the tsnet server and waits for it to be ready.
func (s *Server) Start(ctx context.Context) error {
	if s.started {
		return fmt.Errorf("server already started")
	}

	// Start by getting local client to trigger initialization
	lc, err := s.server.LocalClient()
	if err != nil {
		return fmt.Errorf("failed to get local client: %w", err)
	}

	s.localClient = lc

	// Wait for the server to be ready by checking status
	status, err := s.localClient.StatusWithoutPeers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	// Wait for backend state to be running
	for status.BackendState != ipn.Running.String() {
		s.logger.Debug("waiting for backend state to be running", "current_state", status.BackendState)

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			status, err = s.localClient.StatusWithoutPeers(ctx)
			if err != nil {
				return fmt.Errorf("failed to get status: %w", err)
			}
		}
	}

	// Get our Tailscale IP addresses
	if status.Self != nil && len(status.Self.TailscaleIPs) > 0 {
		for _, ip := range status.Self.TailscaleIPs {
			if ip.Is4() {
				s.tailscaleIP = ip.String()
			} else if ip.Is6() {
				s.tailscaleIPv6 = ip.String()
			}
		}
	}

	s.started = true
	s.logger.Info("Tailnet connection established", "hostname", s.hostname, "tailscale_ipv4", s.tailscaleIP, "tailscale_ipv6", s.tailscaleIPv6)

	return nil
}

// Listen creates a listener on the Tailscale network.
func (s *Server) Listen(network, address string) (net.Listener, error) {
	if !s.started {
		return nil, fmt.Errorf("server not started")
	}

	return s.server.Listen(network, address)
}

// LocalClient returns the Tailscale LocalClient for API access.
func (s *Server) LocalClient() *local.Client {
	return s.localClient
}

// TailscaleIP returns the IPv4 address assigned to this node.
func (s *Server) TailscaleIP() string {
	return s.tailscaleIP
}

// TailscaleIPv6 returns the IPv6 address assigned to this node.
func (s *Server) TailscaleIPv6() string {
	return s.tailscaleIPv6
}

// Hostname returns the hostname of this tsnet node.
func (s *Server) Hostname() string {
	return s.hostname
}

// Status returns the current Tailscale status.
func (s *Server) Status(ctx context.Context) (*ipnstate.Status, error) {
	if !s.started {
		return nil, fmt.Errorf("server not started")
	}
	return s.localClient.Status(ctx)
}

// Close gracefully shuts down the tsnet server.
func (s *Server) Close() error {
	if !s.started {
		return nil
	}

	if err := s.server.Close(); err != nil {
		return fmt.Errorf("failed to close tsnet server: %w", err)
	}

	s.started = false
	s.logger.Info("disconnected from Tailnet")
	return nil
}

// CleanupStateDir removes the state directory.
// This should only be used in development or testing.
func (s *Server) CleanupStateDir() error {
	if s.started {
		return fmt.Errorf("cannot cleanup state directory while server is running")
	}

	s.logger.Warn("removing Tailscale state directory", "path", s.stateDir)

	if err := os.RemoveAll(s.stateDir); err != nil {
		return fmt.Errorf("failed to remove state directory: %w", err)
	}

	return nil
}

// StateDirPath returns the path to the state directory.
func (s *Server) StateDirPath() string {
	return s.stateDir
}

// StateFileExists checks if the state file exists in the state directory.
func (s *Server) StateFileExists() bool {
	stateFile := filepath.Join(s.stateDir, "tailscaled.state")
	_, err := os.Stat(stateFile)
	return err == nil
}
