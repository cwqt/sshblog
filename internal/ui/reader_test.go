package ui

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cwqt/sshblog/internal/config"
	"github.com/cwqt/sshblog/internal/post"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// runeKey builds a KeyMsg for a single rune, matching how Bubble Tea delivers
// ordinary character presses.
func runeKey(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

// TestReaderTopBottomNavigation covers the vim-style `G` (bottom) and `gg`
// (top) shortcuts, including that a lone `g` followed by another key does not
// jump to the top.
func TestReaderTopBottomNavigation(t *testing.T) {
	r := lipgloss.NewRenderer(os.Stdout)
	m := Model{
		Page:     PageReader,
		Ready:    true,
		Width:    80,
		Height:   10,
		Content:  strings.TrimSuffix(strings.Repeat("line\n", 100), "\n"),
		renderer: r,
		styles:   newStyles(r),
		cfg:      config.Default(),
	}
	bottom := m.maxScroll()
	if bottom == 0 {
		t.Fatal("expected a scrollable body")
	}

	// G jumps to the bottom.
	mm, _ := m.updateReader(runeKey('G'))
	if got := mm.(Model).ScrollOffset; got != bottom {
		t.Errorf("G: ScrollOffset = %d, want %d", got, bottom)
	}

	// gg jumps back to the top.
	mm, _ = mm.(Model).updateReader(runeKey('g'))
	if !mm.(Model).pendingG {
		t.Fatal("first g should arm the pending-g state")
	}
	mm, _ = mm.(Model).updateReader(runeKey('g'))
	if got := mm.(Model).ScrollOffset; got != 0 {
		t.Errorf("gg: ScrollOffset = %d, want 0", got)
	}

	// g then a non-g key must not jump to the top.
	atBottom := m // fresh model still at the top
	atBottom.ScrollOffset = bottom
	mm, _ = atBottom.updateReader(runeKey('g'))
	mm, _ = mm.(Model).updateReader(runeKey('j'))
	if got := mm.(Model).ScrollOffset; got == 0 {
		t.Errorf("g then j should not go to top (offset=%d)", got)
	}
}

// TestTimeAgo covers the relative-time bucketing used on the reader date line.
func TestTimeAgo(t *testing.T) {
	now := time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		ago  time.Duration
		want string
	}{
		{30 * time.Second, "just now"},
		{-time.Hour, "just now"}, // future
		{1 * time.Minute, "1 minute ago"},
		{5 * time.Minute, "5 minutes ago"},
		{1 * time.Hour, "1 hour ago"},
		{3 * time.Hour, "3 hours ago"},
		{25 * time.Hour, "1 day ago"},
		{6 * 24 * time.Hour, "6 days ago"},
		{7 * 24 * time.Hour, "1 week ago"},
		{28 * 24 * time.Hour, "4 weeks ago"},
		{60 * 24 * time.Hour, "2 months ago"},
		{400 * 24 * time.Hour, "1 year ago"},
	}
	for _, c := range cases {
		if got := timeAgo(now.Add(-c.ago), now); got != c.want {
			t.Errorf("timeAgo(-%s) = %q, want %q", c.ago, got, c.want)
		}
	}
}

// TestReaderFrameFitsViewport guards against the reader frame overflowing the
// terminal height, which in alt-screen mode scrolls the post title off the top.
func TestReaderFrameFitsViewport(t *testing.T) {
	r := lipgloss.NewRenderer(os.Stdout)
	m := Model{
		reading: post.Post{
			Title: "A Reasonably Long Blog Post Title",
			Date:  time.Now(),
		},
		Page:     PageReader,
		Ready:    true,
		renderer: r,
		styles:   newStyles(r),
		cfg:      config.Default(),
	}

	long := strings.TrimSuffix(strings.Repeat("content line\n", 100), "\n")
	contents := map[string]string{
		"short": "line1\nline2\nline3\nline4\nline5",
		"long":  long,
	}

	for name, content := range contents {
		m.Content = content
		for _, h := range []int{6, 8, 10, 12, 15, 20, 24, 40} {
			m.Height, m.Width, m.ScrollOffset = h, 80, 0
			out := m.viewReader()
			if rh := lipgloss.Height(out); rh > h {
				t.Errorf("[%s] termH=%d: frame overflowed to %d lines", name, h, rh)
			}
			firstLine := strings.SplitN(out, "\n", 2)[0]
			if !strings.Contains(firstLine, "Blog Post Title") {
				t.Errorf("[%s] termH=%d: title not on first line: %q", name, h, firstLine)
			}
		}
	}
}
