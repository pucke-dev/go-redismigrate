package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbletea"

	"github.com/pucke-dev/go-redismigrate/internal/migrate"
	"github.com/pucke-dev/go-redismigrate/internal/stats"
)

type (
	Model struct {
		// migrator handles the migration logic.
		migrator *migrate.Migrator

		// metrics provides migration progress and statistics.
		metrics *stats.Metrics

		// progressBar is the progress bar model used to display migration progress.
		progressBar progress.Model

		// view is the view model that renders the UI.
		view   *View
		err    error
		width  int
		height int
	}

	TickMsg  time.Time
	DoneMsg  struct{}
	ErrorMsg error
)

func NewModel(migrator *migrate.Migrator, metrics *stats.Metrics) Model {
	prog := progress.New(progress.WithDefaultGradient())

	return Model{
		migrator:    migrator,
		metrics:     metrics,
		progressBar: prog,
		view:        NewView(),
	}
}

func (m Model) Init() tea.Cmd {
	// Trigger periodic updates.
	return TickCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progressBar.Width = min(msg.Width-4, 80)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case tea.QuitMsg:
		return m, tea.Quit

	case TickMsg:
		return m, TickCmd()

	case DoneMsg:
		return m, tea.Quit

	case ErrorMsg:
		m.err = msg
		return m, nil
	}

	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return m.view.RenderError(m.err)
	}

	return m.view.RenderMigrationProgress(ViewData{
		Config:      m.migrator.GetConfig(),
		Metrics:     m.metrics,
		ProgressBar: m.progressBar,
	})
}

func TickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func DoneCmd() tea.Cmd {
	return func() tea.Msg {
		return DoneMsg{}
	}
}

func ErrorCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return ErrorMsg(err)
	}
}
