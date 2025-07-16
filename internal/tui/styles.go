package tui

import "github.com/charmbracelet/lipgloss"

var Styles = DefaultStyles()

type StyleSystem struct {
	Header        lipgloss.Style
	ErrorHeader   lipgloss.Style
	ErrorText     lipgloss.Style
	CodeBlock     lipgloss.Style
	Dash          lipgloss.Style
	Comment       lipgloss.Style
	Description   lipgloss.Style
	Argument      lipgloss.Style
	QuotedString  lipgloss.Style
	FlagDefault   lipgloss.Style
	Flag          lipgloss.Style
	Command       lipgloss.Style
	Program       lipgloss.Style
	Help          lipgloss.Style
	ErrorStatus   lipgloss.Style
	SuccessStatus lipgloss.Style
	InfoStatus    lipgloss.Style
	Config        lipgloss.Style
}

func DefaultStyles() StyleSystem {
	return StyleSystem{
		Header:        lipgloss.NewStyle().Foreground(Colorscheme.Title).Bold(true),
		ErrorHeader:   lipgloss.NewStyle().Foreground(Colorscheme.ErrorHeader).Background(Colorscheme.ErrorBackground).Bold(true).Padding(0, 1),
		ErrorText:     lipgloss.NewStyle().Foreground(Colorscheme.ErrorText),
		CodeBlock:     lipgloss.NewStyle().Background(Colorscheme.Codeblock).Foreground(Salt).Padding(0, 1),
		Dash:          lipgloss.NewStyle().Foreground(Colorscheme.Dash),
		Comment:       lipgloss.NewStyle().Foreground(Colorscheme.Comment).Italic(true),
		Config:        lipgloss.NewStyle().Foreground(Colorscheme.Comment),
		Description:   lipgloss.NewStyle().Foreground(Colorscheme.Description),
		Argument:      lipgloss.NewStyle().Foreground(Colorscheme.Argument),
		QuotedString:  lipgloss.NewStyle().Foreground(Colorscheme.QuotedString),
		Flag:          lipgloss.NewStyle().Foreground(Colorscheme.Flag).Bold(true),
		FlagDefault:   lipgloss.NewStyle().Foreground(Colorscheme.FlagDefault),
		Command:       lipgloss.NewStyle().Foreground(Colorscheme.Command).Bold(true),
		Program:       lipgloss.NewStyle().Foreground(Colorscheme.Program).Bold(true),
		Help:          lipgloss.NewStyle().Foreground(Colorscheme.Help).Italic(true).MarginTop(1),
		ErrorStatus:   lipgloss.NewStyle().Foreground(Colorscheme.ErrorHeader),
		SuccessStatus: lipgloss.NewStyle().Foreground(Colorscheme.Success),
		InfoStatus:    lipgloss.NewStyle().Foreground(Colorscheme.Info),
	}
}
