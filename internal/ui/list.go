package ui

import (
	"github.com/cwqt/sshblog/internal/config"
	"github.com/cwqt/sshblog/internal/post"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// item adapts a post to the bubbles/list Item interface. Each row shows the
// title with the date on a second line beneath it; FilterValue is the title so
// the delegate's match highlighting lines up with what the user typed.
type item struct{ p post.Post }

func (i item) Title() string       { return i.p.Title }
func (i item) Description() string { return i.p.Date.Format("2006-01-02") }
func (i item) FilterValue() string { return i.p.Title }

// rssBinding / openBinding surface in the list's help footer only; the keys
// themselves are handled in updateIndex, not by the list.
var (
	openBinding = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open"))
	rssBinding  = key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rss"))
)

// newPostList builds the index's post list. Every style is rebuilt from the
// session renderer r (rather than bubbles' package-level defaults) so colors
// track the connecting client's terminal, matching the rest of the UI — with
// the global renderer the list would render colorless under `make dev`, where
// the server's stdout is a log file.
func newPostList(posts []post.Post, r *lipgloss.Renderer, cfg config.Config) list.Model {
	items := make([]list.Item, len(posts))
	for i, p := range posts {
		items[i] = item{p: p}
	}

	l := list.New(items, newDelegate(r), 0, 0)

	// We render our own logo + description above the list, so hide the list's
	// own title bar and keep the rest of the default chrome.
	l.Title = cfg.Title
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)

	// Rebind styles and the sub-components list.New wired to the defaults.
	l.Styles = newListStyles(r)
	l.FilterInput.PromptStyle = l.Styles.FilterPrompt
	l.FilterInput.Cursor.Style = l.Styles.FilterCursor
	l.Paginator.ActiveDot = l.Styles.ActivePaginationDot.String()
	l.Paginator.InactiveDot = l.Styles.InactivePaginationDot.String()
	l.Help.Styles = newHelpStyles(r)

	// Advertise the RSS key only when a feed is actually configured; otherwise
	// keep it out of the help footer entirely (and updateIndex ignores `r`).
	extra := []key.Binding{openBinding}
	if feedTarget(cfg) != "" {
		extra = append(extra, rssBinding)
	}
	keys := func() []key.Binding { return extra }
	l.AdditionalShortHelpKeys = keys
	l.AdditionalFullHelpKeys = keys

	return l
}

// newDelegate is the default two-line delegate (title over date), themed via r.
func newDelegate(r *lipgloss.Renderer) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.ShowDescription = true
	d.SetHeight(2)
	d.SetSpacing(1)
	d.Styles = newDelegateStyles(r)
	return d
}

// newDelegateStyles mirrors list.NewDefaultItemStyles but binds to r and uses
// the blog's turquoise accent for the selected row.
func newDelegateStyles(r *lipgloss.Renderer) list.DefaultItemStyles {
	var s list.DefaultItemStyles

	s.NormalTitle = r.NewStyle().Foreground(text).Padding(0, 0, 0, 2)
	s.NormalDesc = s.NormalTitle.Foreground(muted)

	s.SelectedTitle = r.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lightTurquoise).
		Foreground(lightTurquoise).
		Bold(true).
		Padding(0, 0, 0, 1)
	s.SelectedDesc = s.SelectedTitle.Foreground(lightTurquoise)

	s.DimmedTitle = r.NewStyle().Foreground(muted).Padding(0, 0, 0, 2)
	s.DimmedDesc = s.DimmedTitle

	s.FilterMatch = r.NewStyle().Underline(true)
	return s
}

// newListStyles mirrors list.DefaultStyles but binds to r and swaps the accent
// colors for the blog's palette.
func newListStyles(r *lipgloss.Renderer) list.Styles {
	verySubdued := lipgloss.AdaptiveColor{Light: "#DDDADA", Dark: "#3C3C3C"}
	subdued := lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}

	var s list.Styles
	s.TitleBar = r.NewStyle().Padding(0, 0, 1, 2)
	s.Title = r.NewStyle().Foreground(text).Bold(true).Padding(0, 1)
	s.Spinner = r.NewStyle().Foreground(muted)
	s.FilterPrompt = r.NewStyle().Foreground(lightTurquoise)
	s.FilterCursor = r.NewStyle().Foreground(hotPink)
	s.DefaultFilterCharacterMatch = r.NewStyle().Underline(true)
	s.StatusBar = r.NewStyle().Foreground(muted).Padding(0, 0, 1, 2)
	s.StatusEmpty = r.NewStyle().Foreground(subdued)
	s.StatusBarActiveFilter = r.NewStyle().Foreground(text)
	s.StatusBarFilterCount = r.NewStyle().Foreground(verySubdued)
	s.NoItems = r.NewStyle().Foreground(muted)
	s.ArabicPagination = r.NewStyle().Foreground(subdued)
	s.PaginationStyle = r.NewStyle().PaddingLeft(2)
	s.HelpStyle = r.NewStyle().Padding(1, 0, 0, 2)
	s.ActivePaginationDot = r.NewStyle().Foreground(lightTurquoise).SetString("•")
	s.InactivePaginationDot = r.NewStyle().Foreground(verySubdued).SetString("•")
	s.DividerDot = r.NewStyle().Foreground(verySubdued).SetString(" • ")
	return s
}

// newHelpStyles binds the help footer's key/description/separator styles to r.
func newHelpStyles(r *lipgloss.Renderer) help.Styles {
	keyColor := lipgloss.AdaptiveColor{Light: "#909090", Dark: "#626262"}
	descColor := lipgloss.AdaptiveColor{Light: "#B2B2B2", Dark: "#4A4A4A"}
	sepColor := lipgloss.AdaptiveColor{Light: "#DDDADA", Dark: "#3C3C3C"}

	var s help.Styles
	s.Ellipsis = r.NewStyle().Foreground(sepColor)
	s.ShortKey = r.NewStyle().Foreground(keyColor)
	s.ShortDesc = r.NewStyle().Foreground(descColor)
	s.ShortSeparator = r.NewStyle().Foreground(sepColor)
	s.FullKey = s.ShortKey
	s.FullDesc = s.ShortDesc
	s.FullSeparator = s.ShortSeparator
	return s
}
