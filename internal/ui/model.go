package ui

import (
	"path"
	"strings"

	"github.com/cwqt/sshblog/internal/config"
	"github.com/cwqt/sshblog/internal/post"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Page represents which page is currently displayed.
type Page int

const (
	PageIndex Page = iota
	PageReader
)

// Model is the Bubble Tea model for the SSH blog app.
type Model struct {
	Page         Page
	ScrollOffset int    // Scroll position in reader
	Width        int    // Terminal width
	Height       int    // Terminal height
	Ready        bool   // Whether window size is known
	Content      string // Rendered markdown content for reader

	list       list.Model // Post index (bubbles/list): navigation, filtering, help
	reading    post.Post  // Post currently open in the reader
	feedTarget string     // URL or path the RSS key hands to the system opener
	openErr    error      // Last feed-open failure, shown under the index list

	renderer *lipgloss.Renderer // Session-bound renderer (client terminal)
	styles   Styles             // Styles built from the renderer
	cfg      config.Config      // User-facing text loaded from sshblog.yaml
}

// New creates a new Model with the given posts and config. The renderer must be
// bound to the SSH session (via bubbletea.MakeRenderer) so colors reflect the
// client's terminal rather than the server's stdout.
func New(posts []post.Post, renderer *lipgloss.Renderer, cfg config.Config) Model {
	return Model{
		Page:       PageIndex,
		list:       newPostList(posts, renderer, cfg),
		feedTarget: feedTarget(cfg),
		renderer:   renderer,
		styles:     newStyles(renderer),
		cfg:        cfg,
	}
}

// feedTarget resolves what the RSS key hands to the system opener. With a
// public URL configured it prefers the hosted feed (served by a co-located web
// server) at "<url>/<basename(feed_path)>"; otherwise it falls back to opening
// the generated file directly, which is mainly useful in local development.
// An empty result means no feed is available.
func feedTarget(cfg config.Config) string {
	if cfg.URL != "" {
		name := "rss.xml"
		if cfg.FeedPath != "" {
			name = path.Base(cfg.FeedPath)
		}
		return strings.TrimRight(cfg.URL, "/") + "/" + name
	}
	return cfg.FeedPath
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Ready = true
		m.resizeList()

		// Re-render content if in reader view
		if m.Page == PageReader {
			m.Content = m.renderMarkdown(m.reading.Content, m.Width)
		}
		return m, nil

	case feedOpenedMsg:
		// Surface any opener failure on its own line under the list, re-sizing
		// the list so the extra line doesn't push the frame off the alt screen.
		m.openErr = msg.err
		m.resizeList()
		return m, nil

	case tea.KeyMsg:
		switch m.Page {
		case PageIndex:
			return m.updateIndex(msg)
		case PageReader:
			return m.updateReader(msg)
		}
	}

	// Forward housekeeping messages (cursor blink, status-message timer, etc.)
	// to the list while the index is active so its own state stays live.
	if m.Page == PageIndex {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	if !m.Ready {
		return "Loading..."
	}

	switch m.Page {
	case PageIndex:
		return m.viewIndex()
	case PageReader:
		return m.viewReader()
	default:
		return ""
	}
}
