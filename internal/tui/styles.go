package tui

import "charm.land/lipgloss/v2"

var (
	accent  = lipgloss.Color("#7B2FBE")
	fg      = lipgloss.Color("#C0CAF5")
	dim     = lipgloss.Color("#565F89")
	success = lipgloss.Color("#9ECE6A")
	errC    = lipgloss.Color("#F7768E")
	info    = lipgloss.Color("#7DCFFF")
)

var baseStyle = lipgloss.NewStyle().Foreground(fg)

var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(accent).
	Padding(0, 1)

var sectionStyle = lipgloss.NewStyle().
	Foreground(accent).
	Bold(true)

var dimStyle = lipgloss.NewStyle().
	Foreground(dim).
	Italic(true)

var accentStyle = lipgloss.NewStyle().
	Foreground(accent).
	Bold(true)

var successStyle = lipgloss.NewStyle().
	Foreground(success).
	Bold(true)

var errorStyle = lipgloss.NewStyle().
	Foreground(errC).
	Bold(true)

var queryStyle = lipgloss.NewStyle().
	Foreground(accent).
	Bold(true).
	Italic(true)

var footerStyle = lipgloss.NewStyle().
	Foreground(dim).
	Padding(0, 1)

var cursorStyle = lipgloss.NewStyle().
	Foreground(fg).
	Background(accent).
	Padding(0, 1)

var inputStyle = lipgloss.NewStyle().
	Foreground(fg).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(accent).
	Padding(0, 1).
	Width(60)

var headerBox = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(accent).
	Padding(1, 2)

var errorBox = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(errC).
	Padding(1, 2).
	Foreground(errC)

var thoughtStyle = lipgloss.NewStyle().
	Foreground(dim).
	Italic(true)

var toolCallStyle = lipgloss.NewStyle().
	Foreground(accent).
	Bold(true)

var separatorStyle = lipgloss.NewStyle().
	Foreground(dim)

var actionStyle = lipgloss.NewStyle().
	Foreground(info).
	Bold(true)

var answerStyle = lipgloss.NewStyle().
	Foreground(fg).
	Bold(true)

var dividerSectionStyle = lipgloss.NewStyle().
	Foreground(dim).
	Bold(true)

var dividerLabelStyle = lipgloss.NewStyle().
	Foreground(accent).
	Bold(true)

type spinnerFrames []string

var dotFrames = spinnerFrames{
	"   ",
	".  ",
	".. ",
	"...",
	" ..",
	"  .",
	"   ",
}

func spinnerFrame(frames spinnerFrames, i int) string {
	return frames[i%len(frames)]
}