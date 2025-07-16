package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	Colorscheme = DefaultColorScheme()

	Charcoal = lipgloss.Color("#1A1B26")
	Ash      = lipgloss.Color("#E8E8E8")
	Salt     = lipgloss.Color("#FFFFFF")
	CodeBG   = lipgloss.Color("#2A2D3A")
	Charple  = lipgloss.Color("#8B6FFF")
	Malibu   = lipgloss.Color("#40B4FF")
	Guppy    = lipgloss.Color("#9090FF")
	Pony     = lipgloss.Color("#FF6FDF")
	Cheeky   = lipgloss.Color("#FF9FE0")
	Squid    = lipgloss.Color("#2D3748")
	Oyster   = lipgloss.Color("#A0A0A0")
	Orange   = lipgloss.Color("#FF9966")
	Guac     = lipgloss.Color("#48D597")
	Coral    = lipgloss.Color("#FF7A9A")
	Salmon   = lipgloss.Color("#FF88AA")
	Butter   = lipgloss.Color("#FFE066")
	Cherry   = lipgloss.Color("#FF5599")
	Smoke    = lipgloss.Color("#8A8A8A")
)

// ColorScheme defines the color palette for the application.
type ColorScheme struct {
	Base            lipgloss.Color
	Title           lipgloss.Color
	Description     lipgloss.Color
	Codeblock       lipgloss.Color
	Program         lipgloss.Color
	DimmedArgument  lipgloss.Color
	Comment         lipgloss.Color
	Flag            lipgloss.Color
	FlagDefault     lipgloss.Color
	Command         lipgloss.Color
	QuotedString    lipgloss.Color
	Argument        lipgloss.Color
	Help            lipgloss.Color
	Dash            lipgloss.Color
	ErrorHeader     lipgloss.Color
	ErrorBackground lipgloss.Color
	ErrorText       lipgloss.Color
	Success         lipgloss.Color
	Warning         lipgloss.Color
	Info            lipgloss.Color
}

func DefaultColorScheme() ColorScheme {
	return ColorScheme{
		Base:            Ash,
		Title:           Charple,
		Description:     Ash,
		Codeblock:       CodeBG,
		Program:         Guppy,
		Command:         Orange,
		DimmedArgument:  Oyster,
		Comment:         Smoke,
		Flag:            Guac,
		FlagDefault:     Smoke,
		QuotedString:    Salmon,
		Argument:        Ash,
		Help:            Smoke,
		Dash:            Oyster,
		ErrorHeader:     Butter,
		ErrorBackground: Cherry,
		ErrorText:       Salt,
		Success:         Guac,
		Warning:         Butter,
		Info:            Guppy,
	}
}
