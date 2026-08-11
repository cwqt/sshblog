// Package config loads the SSH blog's user-facing text from a YAML file,
// falling back to built-in defaults when the file or individual fields are
// absent so the blog always renders something sensible.
package config

import (
	"errors"
	"io/fs"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds the display strings and post source for the TUI.
type Config struct {
	// Title is the logo/header shown at the top of the index.
	Title string `yaml:"title"`
	// Description is the informational block shown under the title.
	Description string `yaml:"description"`
	// Help holds the key-hint footers for each page.
	Help Help `yaml:"help"`
	// Posts is the directory of markdown posts to load.
	Posts string `yaml:"posts"`
	// URL is the site's public base URL (e.g. "https://cass.si"). Optional; it
	// supplies the RSS feed's channel and item permalinks. Left empty, the feed
	// still renders but items carry slug-only (non-permalink) guids.
	URL string `yaml:"url"`
	// FeedPath is where `sshblog -gen-feed` writes the RSS document, typically a
	// path inside a co-located web server's document root (e.g.
	// "/var/www/html/rss.xml") so the feed is served at a stable public URL.
	// Optional; when set together with URL, the RSS key opens
	// "<url>/<basename(feed_path)>", otherwise it opens this file directly.
	FeedPath string `yaml:"feed_path"`
}

// Help holds the footer hint text for each page.
type Help struct {
	Index  string `yaml:"index"`
	Reader string `yaml:"reader"`
}

// Default returns the built-in configuration.
func Default() Config {
	return Config{
		Title: "sshblog",
		Description: "A little blog you read over SSH.\n" +
			"Edit sshblog.yaml to change this text.\n" +
			"Drop markdown posts into the posts directory.",
		Help: Help{
			Index:  "↑/k up • ↓/j down • enter select • q quit",
			Reader: "↑/k up • ↓/j down • esc/backspace back • q quit",
		},
		Posts: "posts",
	}
}

// Load reads configuration from path. A missing file is not an error: the
// built-in defaults are returned. Any field left empty in the file also falls
// back to its default, so a partial config is valid.
func Load(path string) (Config, error) {
	def := Default()

	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return def, nil
	}
	if err != nil {
		return def, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return def, err
	}

	cfg.applyDefaults(def)
	return cfg, nil
}

// applyDefaults fills any empty field from def.
func (c *Config) applyDefaults(def Config) {
	if strings.TrimSpace(c.Title) == "" {
		c.Title = def.Title
	}
	if strings.TrimSpace(c.Description) == "" {
		c.Description = def.Description
	}
	if strings.TrimSpace(c.Help.Index) == "" {
		c.Help.Index = def.Help.Index
	}
	if strings.TrimSpace(c.Help.Reader) == "" {
		c.Help.Reader = def.Help.Reader
	}
	if strings.TrimSpace(c.Posts) == "" {
		c.Posts = def.Posts
	}
	// A YAML block scalar keeps a trailing newline; drop it so it doesn't add
	// a blank line to the rendered description.
	c.Description = strings.TrimRight(c.Description, "\n")
}
