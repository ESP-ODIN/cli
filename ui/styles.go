package ui

import "github.com/charmbracelet/lipgloss"

var (
	colorOrange = lipgloss.Color("#E8540A")
	colorWhite  = lipgloss.Color("#EFEFEF")
	colorGray   = lipgloss.Color("#6E7681")
	colorGreen  = lipgloss.Color("#22C55E")
	colorRed    = lipgloss.Color("#EF4444")
	colorDim    = lipgloss.Color("#3D3D3D")

	StyleAccent = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorOrange)

	StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite)

	StyleName = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorOrange)

	StyleVersion = lipgloss.NewStyle().
			Foreground(colorGray)

	StyleDesc = lipgloss.NewStyle().
			Foreground(colorWhite)

	StyleSuccess = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorGreen)

	StyleError = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorRed)

	StyleMuted = lipgloss.NewStyle().
			Foreground(colorGray)

	StyleDivider = lipgloss.NewStyle().
			Foreground(colorDim)

	StyleStep = lipgloss.NewStyle().
			Foreground(colorGray)

	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorOrange)
)

const Banner = `  ᚩ  ODIN`

const Tagline = "     package manager for AI agents  v0.1.0"
