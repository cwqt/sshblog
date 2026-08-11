// Package feed generates an RSS 2.0 document from the blog's posts and config.
// The SSH TUI has no HTTP server of its own, so the feed is rendered into a
// scrollable "rss.xml" view a reader can inspect or copy; a host that fronts
// the blog over HTTP can serve the same bytes.
package feed

import (
	"encoding/xml"
	"strings"
	"time"

	"github.com/cwqt/sshblog/internal/config"
	"github.com/cwqt/sshblog/internal/post"
)

// pubDateFormat is RSS's RFC 822 date, with a numeric zone as recommended by
// the spec (RFC 822's named zones are ambiguous).
const pubDateFormat = time.RFC1123Z

type rss struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel channel  `xml:"channel"`
}

type channel struct {
	Title         string `xml:"title"`
	Link          string `xml:"link,omitempty"`
	Description   string `xml:"description"`
	LastBuildDate string `xml:"lastBuildDate,omitempty"`
	Items         []item `xml:"item"`
}

type item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link,omitempty"`
	GUID        *guid  `xml:"guid,omitempty"`
	PubDate     string `xml:"pubDate,omitempty"`
	Description string `xml:"description"`
}

type guid struct {
	Value       string `xml:",chardata"`
	IsPermaLink bool   `xml:"isPermaLink,attr"`
}

// Generate builds the RSS 2.0 feed for the given posts and config. When
// cfg.URL is set, items get absolute permalink <link>/<guid> values built from
// "<url>/<slug>"; otherwise the slug is used as a non-permalink guid so each
// item still has a stable identifier. Posts are emitted newest-first, matching
// the order LoadAll returns them in.
func Generate(posts []post.Post, cfg config.Config) (string, error) {
	base := strings.TrimRight(strings.TrimSpace(cfg.URL), "/")

	ch := channel{
		Title:       cfg.Title,
		Link:        base,
		Description: cfg.Description,
		Items:       make([]item, 0, len(posts)),
	}

	// lastBuildDate reflects the newest post; deriving it from content (rather
	// than the wall clock) keeps the feed byte-stable for a given post set.
	if len(posts) > 0 && !posts[0].Date.IsZero() {
		ch.LastBuildDate = posts[0].Date.Format(pubDateFormat)
	}

	for _, p := range posts {
		it := item{
			Title:       p.Title,
			Description: p.Content,
		}
		if !p.Date.IsZero() {
			it.PubDate = p.Date.Format(pubDateFormat)
		}
		if base != "" {
			link := base + "/" + p.Slug
			it.Link = link
			it.GUID = &guid{Value: link, IsPermaLink: true}
		} else if p.Slug != "" {
			it.GUID = &guid{Value: p.Slug, IsPermaLink: false}
		}
		ch.Items = append(ch.Items, it)
	}

	doc := rss{Version: "2.0", Channel: ch}

	out, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(out) + "\n", nil
}
