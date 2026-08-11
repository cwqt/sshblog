package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// copyExpiredMsg clears the "copied" confirmation 2s after the RSS key was
// pressed. gen ties it to the press that scheduled it, so a fresh press
// restarts the timer rather than an earlier one cutting it short.
type copyExpiredMsg struct{ gen int }

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
		// Copy the feed URL to the client's clipboard via OSC 52, written
		// straight to the session output. A launcher can only ever run on the
		// SSH host (a headless remote, usually) and never reaches the reader's
		// machine, whereas OSC 52 is interpreted by their own terminal. Then
		// flash a confirmation over the controls line that clears after 2s.
		m.renderer.Output().Copy(m.feedTarget)
		m.copied = true
		m.copyGen++
		gen := m.copyGen
		return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg {
			return copyExpiredMsg{gen: gen}
		})
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

// resizeList sizes the list to the space left under the header, so the index
// frame always fills the viewport exactly without overflowing the alt screen.
// The version footer is inlined onto the list's own help line (see
// inlineFooter), so it costs no extra row.
func (m *Model) resizeList() {
	reserved := lipgloss.Height(m.indexHeader()) + 1 // header + its separator
	h := m.Height - reserved
	if h < 1 {
		h = 1
	}
	m.list.SetSize(m.Width, h)
}

func (m Model) viewIndex() string {
	return m.indexHeader() + "\n" + m.inlineFooter(m.list.View())
}

// inlineFooter right-aligns the version link onto the list's help line (the
// last line of body) when there is room. If it wouldn't fit it is omitted
// rather than wrapped, so the frame height never changes. While a feed copy is
// fresh, the controls text on that line is swapped for the copy confirmation
// for 2s (the version link still trails on the right).
func (m Model) inlineFooter(body string) string {
	lines := strings.Split(body, "\n")
	i := len(lines) - 1
	if m.copied {
		lines[i] = m.styles.copyNotice.Render("Copied RSS link (" + m.feedTarget + ") to clipboard")
	}
	if m.footer == "" || m.Width == 0 {
		return strings.Join(lines, "\n")
	}
	gap := m.Width - lipgloss.Width(lines[i]) - m.footerW
	if gap < 1 {
		return strings.Join(lines, "\n")
	}
	lines[i] += strings.Repeat(" ", gap) + m.footer
	return strings.Join(lines, "\n")
}
