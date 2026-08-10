package ui

import "github.com/charmbracelet/lipgloss"

// Oxocarbon color palette with dark/light mode support
var (
	// Muted text (description, dates, help) - needs to be readable
	muted = lipgloss.AdaptiveColor{Light: "#525252", Dark: "#5e5d5d"}
	// Normal body text
	text = lipgloss.AdaptiveColor{Light: "#262626", Dark: "#f4f4f4"}

	// Accent colors
	lightTurquoise = lipgloss.AdaptiveColor{Light: "#08bdba", Dark: "#3ddbd9"}
	hotPink        = lipgloss.AdaptiveColor{Light: "#a2191f", Dark: "#ee5396"}
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
}

// newStyles builds the style set from a session-bound renderer.
func newStyles(r *lipgloss.Renderer) Styles {
	return Styles{
		// Logo style - bold text color
		logo: r.NewStyle().
			Foreground(text).
			Bold(true).
			PaddingLeft(1),

		// Description styles - muted, informational
		description: r.NewStyle().
			Foreground(muted).
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

		meta: r.NewStyle().
			Foreground(muted).
			MarginBottom(1),

		// Help text - subtle like comments
		help: r.NewStyle().
			Foreground(muted).
			Italic(true).
			MarginTop(1),
	}
}
