package feed

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/cwqt/sshblog/internal/config"
	"github.com/cwqt/sshblog/internal/post"
)

func samplePosts() []post.Post {
	return []post.Post{
		{
			Title:   "Newest & shiniest",
			Date:    time.Date(2026, 2, 1, 9, 0, 0, 0, time.UTC),
			Content: "Body with <angle> brackets.",
			Slug:    "newest",
		},
		{
			Title:   "Older post",
			Date:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			Content: "Older body.",
			Slug:    "older",
		},
	}
}

// TestGenerateIsWellFormedAndOrdered checks the feed parses back to RSS 2.0,
// preserves newest-first order, and escapes special characters.
func TestGenerateIsWellFormedAndOrdered(t *testing.T) {
	cfg := config.Default()
	cfg.URL = "https://cass.si/"

	out, err := Generate(samplePosts(), cfg)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(out, xml.Header) {
		t.Errorf("feed missing XML header:\n%s", out)
	}

	var doc rss
	if err := xml.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("feed is not well-formed XML: %v\n%s", err, out)
	}
	if doc.Version != "2.0" {
		t.Errorf("version = %q, want 2.0", doc.Version)
	}
	if got := len(doc.Channel.Items); got != 2 {
		t.Fatalf("item count = %d, want 2", got)
	}
	if doc.Channel.Items[0].Title != "Newest & shiniest" {
		t.Errorf("items not newest-first: %q", doc.Channel.Items[0].Title)
	}

	// Trailing slash on the base URL must not double up in permalinks.
	if got, want := doc.Channel.Items[0].Link, "https://cass.si/newest"; got != want {
		t.Errorf("item link = %q, want %q", got, want)
	}
	if g := doc.Channel.Items[0].GUID; g == nil || !g.IsPermaLink {
		t.Errorf("permalink guid not set with a base URL: %+v", g)
	}

	// The raw document must escape content, not leak literal angle brackets.
	if strings.Contains(out, "<angle>") {
		t.Errorf("content angle brackets not escaped:\n%s", out)
	}
}

// TestGenerateWithoutURL falls back to slug-only, non-permalink guids.
func TestGenerateWithoutURL(t *testing.T) {
	cfg := config.Default() // URL empty

	out, err := Generate(samplePosts(), cfg)
	if err != nil {
		t.Fatal(err)
	}

	var doc rss
	if err := xml.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("feed is not well-formed XML: %v", err)
	}
	it := doc.Channel.Items[0]
	if it.Link != "" {
		t.Errorf("item link should be empty without a base URL, got %q", it.Link)
	}
	if it.GUID == nil || it.GUID.Value != "newest" || it.GUID.IsPermaLink {
		t.Errorf("expected slug-only non-permalink guid, got %+v", it.GUID)
	}
}

// TestGenerateEmptyPosts still produces a valid channel.
func TestGenerateEmptyPosts(t *testing.T) {
	out, err := Generate(nil, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	var doc rss
	if err := xml.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("empty feed is not well-formed: %v\n%s", err, out)
	}
	if len(doc.Channel.Items) != 0 {
		t.Errorf("expected no items, got %d", len(doc.Channel.Items))
	}
}
