package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateIndex(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// While the filter input is focused, every keystroke belongs to the list.
	if m.list.FilterState() == list.Filtering {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "enter":
		// Open the selected post in the reader.
		if it, ok := m.list.SelectedItem().(item); ok {
			m.reading = it.p
			m.Page = PageReader
			m.ScrollOffset = 0
			m.Content = m.renderMarkdown(it.p.Content, m.Width)
			return m, tea.ClearScreen
		}
		return m, nil

	case "r":
		// No feed configured: `r` isn't advertised in the footer, so ignore it
		// (fall through to the list, which does nothing with it).
		if m.feedTarget == "" {
			break
		}
		// Hand the hosted feed URL (or local file) to the system opener. Clear
		// any prior failure and reclaim its footer line until it reports back.
		if m.openErr != nil {
			m.openErr = nil
			m.resizeList()
		}
		return m, openFeed(m.feedTarget)
	}

	// Everything else — navigation, filtering, pagination, quit — is the
	// list's job.
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// indexHeader is the logo + description block rendered above the post list.
func (m Model) indexHeader() string {
	return m.styles.logo.Render(m.cfg.Title) + "\n\n" +
		m.styles.description.Render(m.cfg.Description)
}

// indexFooter is the optional error line shown under the list (empty when
// there's nothing to report).
func (m Model) indexFooter() string {
	if m.openErr == nil {
		return ""
	}
	return m.styles.help.Render("couldn't open RSS feed: " + m.openErr.Error())
}

// resizeList sizes the list to the space left under the header and above the
// error line, so the index frame always fills the viewport exactly without
// overflowing the alt screen. The version footer is inlined onto the list's
// own help line (see inlineFooter), so it costs no extra row.
func (m *Model) resizeList() {
	reserved := lipgloss.Height(m.indexHeader()) + 1 // header + its separator
	if f := m.indexFooter(); f != "" {
		reserved += lipgloss.Height(f) + 1 // error line + its separator
	}
	h := m.Height - reserved
	if h < 1 {
		h = 1
	}
	m.list.SetSize(m.Width, h)
}

func (m Model) viewIndex() string {
	frame := m.indexHeader() + "\n" + m.inlineFooter(m.list.View())
	if f := m.indexFooter(); f != "" {
		frame += "\n" + f
	}
	return frame
}

// inlineFooter right-aligns the version link onto the list's help line (the
// last line of body) when there is room. If it wouldn't fit it is omitted
// rather than wrapped, so the frame height never changes.
func (m Model) inlineFooter(body string) string {
	if m.footer == "" || m.Width == 0 {
		return body
	}
	lines := strings.Split(body, "\n")
	i := len(lines) - 1
	gap := m.Width - lipgloss.Width(lines[i]) - m.footerW
	if gap < 1 {
		return body
	}
	lines[i] += strings.Repeat(" ", gap) + m.footer
	return strings.Join(lines, "\n")
}
