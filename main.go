package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cwqt/sshblog/internal/config"
	"github.com/cwqt/sshblog/internal/post"
	"github.com/cwqt/sshblog/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/ssh"
)

const (
	host       = "0.0.0.0"
	port       = "2222"
	configPath = "sshblog.yaml"
)

func main() {
	// Load config (falls back to defaults if sshblog.yaml is absent)
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Printf("Config warning (using defaults where needed): %v", err)
	}

	// Load posts from the configured directory
	posts, err := post.LoadAll(cfg.Posts)
	if err != nil {
		log.Fatalf("Failed to load posts: %v", err)
	}
	log.Printf("Loaded %d posts from %q", len(posts), cfg.Posts)

	// Create SSH server
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler(posts, cfg)),
		),
	)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Handle shutdown gracefully
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Starting SSH server on %s:%s", host, port)
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-done
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown error: %v", err)
	}
}

func teaHandler(posts []post.Post, cfg config.Config) bubbletea.Handler {
	return func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		// Build a renderer bound to this session so colors reflect the
		// client's terminal, not the server process's stdout.
		renderer := bubbletea.MakeRenderer(s)
		return ui.New(posts, renderer, cfg), []tea.ProgramOption{tea.WithAltScreen()}
	}
}
