package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateIndex(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}

	case "down", "j":
		if m.Cursor < len(m.Posts)-1 {
			m.Cursor++
		}

	case "enter":
		if len(m.Posts) > 0 && m.Cursor < len(m.Posts) {
			m.Page = PageReader
			m.ScrollOffset = 0
			m.Content = m.renderMarkdown(m.Posts[m.Cursor].Content, m.Width)
			return m, tea.ClearScreen
		}
	}

	return m, nil
}

func (m Model) viewIndex() string {
	var b strings.Builder

	// Header
	b.WriteString(m.styles.logo.Render(m.cfg.Title))
	b.WriteString("\n\n")

	// Description
	b.WriteString(m.styles.description.Render(m.cfg.Description))
	b.WriteString("\n\n")

	// Post list
	for i, p := range m.Posts {
		date := m.styles.date.Render(p.Date.Format("2006-01-02"))

		var title string
		if i == m.Cursor {
			title = m.styles.selected.Render("> " + p.Title)
		} else {
			title = m.styles.normal.Render("  " + p.Title)
		}

		b.WriteString(fmt.Sprintf("%s %s\n", date, title))
	}

	// Help text
	b.WriteString(m.styles.help.Render(m.cfg.Help.Index))

	return b.String()
}
