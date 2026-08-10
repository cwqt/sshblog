package ui

import (
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
)

func (m Model) updateReader(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "esc", "backspace":
		m.Page = PageIndex
		m.ScrollOffset = 0
		return m, tea.ClearScreen

	case "up", "k":
		if m.ScrollOffset > 0 {
			m.ScrollOffset--
		}

	case "down", "j":
		m.ScrollOffset++
		// Clamp to max scroll
		m.ScrollOffset = m.clampScroll(m.ScrollOffset)
	}

	return m, nil
}

func (m Model) clampScroll(offset int) int {
	lines := strings.Split(m.Content, "\n")
	maxScroll := len(lines) - m.readerVisibleHeight()
	if maxScroll < 0 {
		maxScroll = 0
	}
	if offset > maxScroll {
		return maxScroll
	}
	return offset
}

// readerHeader renders the title + date block shown above the post body.
func (m Model) readerHeader() string {
	post := m.Posts[m.Cursor]
	return m.styles.header.Render(post.Title) + "\n" +
		m.styles.meta.Render(post.Date.Format("January 2, 2006"))
}

// readerVisibleHeight is how many body lines fit given the current viewport,
// measuring the actual header/help chrome rather than guessing a fixed count.
// Layout is: header + blank separator + body + help.
func (m Model) readerVisibleHeight() int {
	chrome := lipgloss.Height(m.readerHeader()) +
		lipgloss.Height(m.styles.help.Render(m.cfg.Help.Reader)) + 1
	vh := m.Height - chrome
	if vh < 1 {
		vh = 1
	}
	return vh
}

func (m Model) viewReader() string {
	if m.Cursor >= len(m.Posts) {
		return "No post selected"
	}

	// Content with scrolling
	lines := strings.Split(m.Content, "\n")
	visibleHeight := m.readerVisibleHeight()

	// Clamp scroll offset
	maxScroll := len(lines) - visibleHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.ScrollOffset > maxScroll {
		m.ScrollOffset = maxScroll
	}

	// Show visible lines
	endLine := m.ScrollOffset + visibleHeight
	if endLine > len(lines) {
		endLine = len(lines)
	}
	body := strings.Join(lines[m.ScrollOffset:endLine], "\n")

	// header + blank separator + body + help
	frame := m.readerHeader() + "\n" + body + "\n" + m.styles.help.Render(m.cfg.Help.Reader)

	// Safety net: never emit more lines than the viewport, or the alt screen
	// scrolls the title off the top. Trim from the bottom, keeping the title.
	if m.Height > 0 {
		if fl := strings.Split(frame, "\n"); len(fl) > m.Height {
			frame = strings.Join(fl[:m.Height], "\n")
		}
	}

	return frame
}

// maxRenderWidth caps line length for readability on wide terminals.
const maxRenderWidth = 80

func (m Model) renderMarkdown(content string, width int) string {
	if width <= 0 || width > maxRenderWidth {
		width = maxRenderWidth
	}

	// Glamour's own word-wrap pads every line to the wrap width and its
	// document margin fights the wrap, producing trailing whitespace and
	// stray one-word lines. Instead: collapse the source's soft line breaks,
	// let glamour style without wrapping/margin, then wrap ANSI-aware here.
	opts := []glamour.TermRendererOption{
		glamour.WithStyles(glamourStyle(m.renderer)),
		glamour.WithWordWrap(0),
	}
	// Drive glamour's colors from the session's color profile.
	if m.renderer != nil {
		opts = append(opts, glamour.WithColorProfile(m.renderer.ColorProfile()))
	}

	r, err := glamour.NewTermRenderer(opts...)
	if err != nil {
		return content
	}

	rendered, err := r.Render(collapseSoftBreaks(content))
	if err != nil {
		return content
	}

	// Wrap on word boundaries (only breaking a word when it alone exceeds the
	// limit, e.g. a long URL); the Hardwrap clamps the couple of columns Wrap
	// can leave hanging at a breakpoint so no line ever exceeds the cap.
	rendered = xansi.Hardwrap(xansi.Wrap(strings.TrimSpace(rendered), width, ""), width, false)

	// Trim any space left at a wrap boundary.
	lines := strings.Split(rendered, "\n")
	for i, ln := range lines {
		lines[i] = strings.TrimRight(ln, " ")
	}
	return strings.Join(lines, "\n")
}

// glamourStyle returns the dark or light glamour style (matching the session's
// detected background) with the document margin/indent removed so our own
// wrapping controls the layout.
func glamourStyle(r *lipgloss.Renderer) ansi.StyleConfig {
	cfg := styles.LightStyleConfig
	if r != nil && r.HasDarkBackground() {
		cfg = styles.DarkStyleConfig
	}
	zero := uint(0)
	cfg.Document.Margin = &zero
	cfg.Document.Indent = &zero
	return cfg
}

var blockPrefix = regexp.MustCompile(`^\s*([-*+>#]|\d+\.|\|)`)

// collapseSoftBreaks joins the soft (single-newline) line breaks inside prose
// paragraphs into spaces so the text can be re-wrapped cleanly, while leaving
// blank lines, fenced code blocks, and block-level markdown (lists, headings,
// quotes, tables) untouched.
func collapseSoftBreaks(md string) string {
	lines := strings.Split(md, "\n")
	out := make([]string, 0, len(lines))
	inFence := false
	for _, ln := range lines {
		trimmed := strings.TrimSpace(ln)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			out = append(out, ln)
			continue
		}
		prev := ""
		if len(out) > 0 {
			prev = out[len(out)-1]
		}
		// Keep a line separate when it (or the previous line) is code, blank,
		// or block-level markdown that must not be joined.
		if inFence || trimmed == "" || blockPrefix.MatchString(ln) ||
			prev == "" || strings.TrimSpace(prev) == "" ||
			blockPrefix.MatchString(prev) ||
			strings.HasPrefix(strings.TrimSpace(prev), "```") {
			out = append(out, ln)
			continue
		}
		out[len(out)-1] = strings.TrimRight(prev, " ") + " " + trimmed
	}
	return strings.Join(out, "\n")
}
