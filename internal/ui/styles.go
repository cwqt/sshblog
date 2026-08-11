package ui

import "github.com/charmbracelet/lipgloss"

// Oxocarbon color palette with dark/light mode support
var (
	// Muted text (dates, help) - needs to be readable. The dark value is a
	// mid-grey (Carbon gray-50) rather than a near-black grey so the subtext
	// keeps enough contrast against a dark terminal background.
	muted = lipgloss.AdaptiveColor{Light: "#525252", Dark: "#8d8d8d"}
	// Description block - a touch lighter than muted in dark mode (Carbon
	// gray-40) since it's the main informational text under the title.
	descriptionColor = lipgloss.AdaptiveColor{Light: "#525252", Dark: "#a8a8a8"}
	// Normal body text
	text = lipgloss.AdaptiveColor{Light: "#262626", Dark: "#f4f4f4"}

	// Accent colors
	lightTurquoise = lipgloss.AdaptiveColor{Light: "#08bdba", Dark: "#3ddbd9"}
	hotPink        = lipgloss.AdaptiveColor{Light: "#a2191f", Dark: "#ee5396"}

	// Help footer palette, shared by the list's help styles and the version
	// tag inlined onto the same line so they read as one footer. Dark values
	// step down a Carbon gray ramp (50/60/70) so the footer stays legible
	// against a dark background rather than fading into it.
	helpKeyColor  = lipgloss.AdaptiveColor{Light: "#909090", Dark: "#8d8d8d"}
	helpDescColor = lipgloss.AdaptiveColor{Light: "#B2B2B2", Dark: "#6f6f6f"}
	helpSepColor  = lipgloss.AdaptiveColor{Light: "#DDDADA", Dark: "#525252"}
)

// Styles holds all the lipgloss styles for a single session. They are built
// from a session-bound renderer so that color-profile and light/dark
// background detection reflect the connecting SSH client's terminal rather
// than the server process's stdout.
type Styles struct {
	logo     lipgloss.Style
	description  lipgloss.Style
	title    lipgloss.Style
	selected lipgloss.Style
	normal   lipgloss.Style
	date     lipgloss.Style
	header   lipgloss.Style
	meta     lipgloss.Style
	help     lipgloss.Style
	footer   lipgloss.Style
}

// newStyles builds the style set from a session-bound renderer.
func newStyles(r *lipgloss.Renderer) Styles {
	return Styles{
		// Logo style - bold text color
		logo: r.NewStyle().
			Foreground(text).
			Bold(true).
			PaddingLeft(1),

		// Description styles - informational, slightly brighter than muted
		description: r.NewStyle().
			Foreground(descriptionColor).
			MarginBottom(1).
			PaddingLeft(1),

		// Title styles - uses pink for emphasis like markdown headers
		title: r.NewStyle().
			Bold(true).
			Foreground(hotPink).
			MarginBottom(1),

		// Post list styles
		selected: r.NewStyle().
			Foreground(lightTurquoise).
			Bold(true),

		normal: r.NewStyle().
			Foreground(text),

		date: r.NewStyle().
			Foreground(muted).
			Width(12),

		// Reader styles
		header: r.NewStyle().
			Bold(true).
			Foreground(hotPink).
			MarginBottom(1),

		// Reader date/meta line - matches the index description tone
		meta: r.NewStyle().
			Foreground(descriptionColor).
			MarginBottom(1),

		// Help text - subtle like comments
		help: r.NewStyle().
			Foreground(muted).
			Italic(true).
			MarginTop(1),

		// Footer tag (project name + version link), inlined onto the help line
		// and matching its key-hint color so the two read as one footer.
		footer: r.NewStyle().
			Foreground(helpKeyColor),
	}
}
