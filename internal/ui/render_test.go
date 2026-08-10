package ui

import (
	"os"
	"strings"
	"testing"

	"github.com/cwqt/sshblog/internal/post"
	"github.com/charmbracelet/lipgloss"
)

// TestRenderMarkdownWrapping guards the reader's markdown rendering against the
// glamour word-wrap defects: trailing-space padding, and lines that overflow
// the wrap width (which the terminal then soft-wraps mid-word). It also checks
// that rendering never exceeds maxRenderWidth on wide terminals.
func TestRenderMarkdownWrapping(t *testing.T) {
	posts, err := post.LoadAll("../../_posts")
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) == 0 {
		t.Skip("no posts to render")
	}

	r := lipgloss.NewRenderer(os.Stdout)
	m := Model{renderer: r, styles: newStyles(r)}

	for _, width := range []int{60, 80, 100} {
		cap := width
		if cap > maxRenderWidth {
			cap = maxRenderWidth
		}
		for _, p := range posts {
			out := m.renderMarkdown(p.Content, width)
			for i, ln := range strings.Split(out, "\n") {
				if w := lipgloss.Width(ln); w > cap {
					t.Errorf("post %q w=%d: line %d exceeds cap %d (%d cols): %q",
						p.Title, width, i, cap, w, ln)
				}
				if trailing := len(ln) - len(strings.TrimRight(ln, " ")); trailing > 0 {
					t.Errorf("post %q w=%d: line %d has %d trailing spaces: %q",
						p.Title, width, i, trailing, ln)
				}
			}
		}
	}
}

// TestCollapseSoftBreaks verifies prose is joined while structure is preserved.
func TestCollapseSoftBreaks(t *testing.T) {
	in := strings.Join([]string{
		"First prose line",
		"that continues here.",
		"",
		"- a list item",
		"- another item",
		"",
		"```",
		"code line one",
		"code line two",
		"```",
		"",
		"# A heading",
		"body after heading",
	}, "\n")

	got := collapseSoftBreaks(in)

	if !strings.Contains(got, "First prose line that continues here.") {
		t.Errorf("prose soft breaks not collapsed:\n%s", got)
	}
	if !strings.Contains(got, "- a list item\n- another item") {
		t.Errorf("list items must not be joined:\n%s", got)
	}
	if !strings.Contains(got, "code line one\ncode line two") {
		t.Errorf("code block lines must not be joined:\n%s", got)
	}
	if !strings.Contains(got, "# A heading\nbody after heading") {
		t.Errorf("heading must not merge with body:\n%s", got)
	}
}
