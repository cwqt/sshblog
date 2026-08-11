package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cwqt/sshblog/internal/config"
	"github.com/cwqt/sshblog/internal/feed"
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
	genFeed := flag.Bool("gen-feed", false,
		"write the RSS feed to the configured feed_path and exit (build-time generation)")
	flag.Parse()

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

	// Build-time generation: emit the feed to disk (typically a co-located web
	// server's document root) and exit, without starting the SSH server.
	if *genFeed {
		if err := writeFeed(posts, cfg); err != nil {
			log.Fatalf("Failed to generate feed: %v", err)
		}
		log.Printf("Wrote RSS feed to %q", cfg.FeedPath)
		return
	}

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

// writeFeed generates the RSS 2.0 feed for the loaded posts and writes it to
// cfg.FeedPath, which must be set.
func writeFeed(posts []post.Post, cfg config.Config) error {
	if cfg.FeedPath == "" {
		return errors.New("feed_path is not set in " + configPath)
	}
	xml, err := feed.Generate(posts, cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(cfg.FeedPath, []byte(xml), 0o644)
}
