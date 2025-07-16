package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"

	"github.com/pucke-dev/go-redismigrate/internal/migrate"
	"github.com/pucke-dev/go-redismigrate/internal/stats"
)

type (
	ViewData struct {
		Config      migrate.Config
		Metrics     *stats.Metrics
		ProgressBar progress.Model
	}

	View struct{}
)

func NewView() *View {
	return &View{}
}

func (v *View) RenderError(err error) string {
	return Styles.ErrorStatus.Render(fmt.Sprintf("Error: %v", err))
}

// RenderMigrationProgress renders the main migration progress interface.
func (v *View) RenderMigrationProgress(data ViewData) string {
	var content strings.Builder

	content.WriteString(v.renderHeader())
	content.WriteString(v.renderConfiguration(data.Config))
	content.WriteString(v.renderStatus(data.Metrics))
	content.WriteString(v.renderProgress(data.Metrics, data.ProgressBar))
	content.WriteString(v.renderStatistics(data.Metrics, data.Config.Conflict.String()))
	content.WriteString(v.renderPerformance(data.Metrics))
	content.WriteString(v.renderHelp())

	return content.String()
}

func (v *View) renderHeader() string {
	var content strings.Builder

	content.WriteString(Styles.Header.Render("Redis Migration Tool"))
	content.WriteString("\n")

	return content.String()
}

func (v *View) renderConfiguration(config migrate.Config) string {
	var content strings.Builder

	content.WriteString("Mode: ")
	content.WriteString(Styles.Flag.Render(config.Mode.String()))
	content.WriteString(" | Pattern: ")
	content.WriteString(Styles.Flag.Render(config.Pattern))
	content.WriteString(" | Conflict: ")
	content.WriteString(Styles.Flag.Render(config.Conflict.String()))
	content.WriteString("\n")

	content.WriteString("Source: ")
	content.WriteString(Styles.Config.Render(config.SourceURL))
	content.WriteString(" | Destination: ")
	content.WriteString(Styles.Config.Render(config.DestURL))
	content.WriteString("\n\n")

	return content.String()
}

func (v *View) renderStatus(metrics *stats.Metrics) string {
	total := metrics.GetTotalKeys()
	processed := metrics.GetProcessedKeys()
	
	switch {
	case total == 0:
		return Styles.Header.Render("Status: No keys to process") + "\n"

	case processed < total:
		return Styles.Header.Render("Status: Processing keys...") + "\n"

	default:
		// As soon as we finish the processing we quit the TUI and display a summary.
		// So we shouldn't reach this point in the TUI.
		return ""
	}
}

func (v *View) renderProgress(metrics *stats.Metrics, progressBar progress.Model) string {
	var content strings.Builder

	content.WriteString(progressBar.ViewAs(metrics.GetProgress()))
	content.WriteString("\n\n")

	return content.String()
}

func (v *View) renderStatistics(metrics *stats.Metrics, conflictMode string) string {
	var content strings.Builder

	content.WriteString(Styles.Header.Render("Total: "))
	content.WriteString(Styles.InfoStatus.Render(FormatCount(metrics.GetTotalKeys())))
	content.WriteString(" | Processed: ")
	content.WriteString(Styles.InfoStatus.Render(FormatCount(metrics.GetProcessedKeys())))
	content.WriteString(" | Success: ")
	content.WriteString(Styles.SuccessStatus.Render(FormatCount(metrics.GetSuccessfulKeys())))
	content.WriteString(" | Failed: ")
	content.WriteString(Styles.ErrorStatus.Render(FormatCount(metrics.GetFailedKeys())))

	switch conflictMode {
	case "skip":
		content.WriteString(" | Skipped: ")
		content.WriteString(Styles.InfoStatus.Render(FormatCount(metrics.GetSkippedKeys())))
	case "overwrite":
		content.WriteString(" | Overwritten: ")
		content.WriteString(Styles.InfoStatus.Render(FormatCount(metrics.GetOverwrittenKeys())))
	}

	content.WriteString("\n")

	return content.String()
}

func (v *View) renderPerformance(metrics *stats.Metrics) string {
	var content strings.Builder

	content.WriteString(Styles.Header.Render("Rate: "))
	content.WriteString(Styles.InfoStatus.Render(FormatRate(metrics.GetProcessingRate())))
	content.WriteString(" | Elapsed: ")
	content.WriteString(Styles.InfoStatus.Render(FormatDuration(metrics.GetElapsed())))

	eta := metrics.GetETA()
	if eta > 0 {
		content.WriteString(" | ETA: ")
		content.WriteString(Styles.InfoStatus.Render(FormatDuration(eta)))
	}

	return content.String()
}

func (v *View) renderHelp() string {
	var content strings.Builder

	content.WriteString("\n")
	content.WriteString(Styles.Help.Render("Press q or Ctrl+C to quit"))

	return content.String()
}
