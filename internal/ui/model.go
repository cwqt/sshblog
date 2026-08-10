package ui

import (
	"github.com/cwqt/sshblog/internal/config"
	"github.com/cwqt/sshblog/internal/post"

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
	Posts        []post.Post
	Page         Page
	Cursor       int    // Index of selected post
	ScrollOffset int    // Scroll position in reader
	Width        int    // Terminal width
	Height       int    // Terminal height
	Ready        bool   // Whether window size is known
	Content      string // Rendered markdown content for reader

	renderer *lipgloss.Renderer // Session-bound renderer (client terminal)
	styles   Styles             // Styles built from the renderer
	cfg      config.Config      // User-facing text loaded from sshblog.yaml
}

// New creates a new Model with the given posts and config. The renderer must be
// bound to the SSH session (via bubbletea.MakeRenderer) so colors reflect the
// client's terminal rather than the server's stdout.
func New(posts []post.Post, renderer *lipgloss.Renderer, cfg config.Config) Model {
	return Model{
		Posts:    posts,
		Page:     PageIndex,
		Cursor:   0,
		renderer: renderer,
		styles:   newStyles(renderer),
		cfg:      cfg,
	}
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

		// Re-render content if in reader view
		if m.Page == PageReader && m.Cursor < len(m.Posts) {
			m.Content = m.renderMarkdown(m.Posts[m.Cursor].Content, m.Width)
		}
		return m, nil

	case tea.KeyMsg:
		switch m.Page {
		case PageIndex:
			return m.updateIndex(msg)
		case PageReader:
			return m.updateReader(msg)
		}
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
