package post

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Post represents a blog post with frontmatter metadata and content.
type Post struct {
	Title   string
	Date    time.Time
	Content string
	Slug    string
}

// Frontmatter represents the YAML frontmatter in markdown files.
type Frontmatter struct {
	Title string    `yaml:"title"`
	Date  time.Time `yaml:"date"`
}

// LoadAll loads all posts from the given directory, sorted by date descending.
func LoadAll(dir string) ([]Post, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return nil, err
	}

	var posts []Post
	for _, file := range files {
		post, err := Load(file)
		if err != nil {
			continue // Skip invalid posts
		}
		posts = append(posts, post)
	}

	// Sort by date descending (newest first)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	return posts, nil
}

// Load parses a single markdown file with YAML frontmatter.
func Load(path string) (Post, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Post{}, err
	}

	frontmatter, content, err := parseFrontmatter(data)
	if err != nil {
		return Post{}, err
	}

	// Extract slug from filename (e.g., "2026-01-03-the-weavers-who-prosper.md" -> "the-weavers-who-prosper")
	base := filepath.Base(path)
	slug := strings.TrimSuffix(base, ".md")
	if len(slug) > 11 {
		slug = slug[11:] // Remove date prefix "YYYY-MM-DD-"
	}

	return Post{
		Title:   frontmatter.Title,
		Date:    frontmatter.Date,
		Content: content,
		Slug:    slug,
	}, nil
}

// parseFrontmatter extracts YAML frontmatter and content from markdown.
func parseFrontmatter(data []byte) (Frontmatter, string, error) {
	var fm Frontmatter

	// Check for frontmatter delimiter
	if !bytes.HasPrefix(data, []byte("---\n")) {
		return fm, string(data), nil
	}

	// Find closing delimiter
	rest := data[4:]
	idx := bytes.Index(rest, []byte("\n---\n"))
	if idx == -1 {
		return fm, string(data), nil
	}

	// Parse YAML
	if err := yaml.Unmarshal(rest[:idx], &fm); err != nil {
		return fm, "", err
	}

	// Content is everything after the closing delimiter
	content := string(rest[idx+5:])
	return fm, strings.TrimSpace(content), nil
}
