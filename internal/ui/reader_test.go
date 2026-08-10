package ui

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cwqt/sshblog/internal/config"
	"github.com/cwqt/sshblog/internal/post"
	"github.com/charmbracelet/lipgloss"
)

// TestReaderFrameFitsViewport guards against the reader frame overflowing the
// terminal height, which in alt-screen mode scrolls the post title off the top.
func TestReaderFrameFitsViewport(t *testing.T) {
	r := lipgloss.NewRenderer(os.Stdout)
	m := Model{
		Posts: []post.Post{{
			Title: "A Reasonably Long Blog Post Title",
			Date:  time.Now(),
		}},
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
