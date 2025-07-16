package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/pucke-dev/go-redismigrate/internal/stats"
)

// FormatCount formats a count value for display.
func FormatCount(count int64) string {
	return fmt.Sprintf("%d", count)
}

// FormatRate formats a processing rate for display.
func FormatRate(rate float64) string {
	return fmt.Sprintf("%.1f keys/sec", rate)
}

// FormatDuration formats a duration for display.
func FormatDuration(duration time.Duration) string {
	return duration.Round(time.Second).String()
}

// FormatUsage returns a beautifully formatted usage message.
func FormatUsage() string {
	usage := strings.Builder{}

	// Title.
	usage.WriteString(Styles.Header.Render("Redis Migration Tool"))
	usage.WriteString("\n\n")

	// Description.
	usage.WriteString(Styles.Description.Render("Migrate Redis keys between instances with real-time progress monitoring"))
	usage.WriteString("\n\n")

	// Usage line.
	usage.WriteString(Styles.Help.Render("Usage:"))
	usage.WriteString(" ")
	usage.WriteString(Styles.Program.Render("redismigrate"))
	usage.WriteString(" ")
	usage.WriteString(Styles.Argument.Render("[options]"))
	usage.WriteString("\n\n")

	// Options header.
	usage.WriteString(Styles.Help.Render("Options:"))
	usage.WriteString("\n")

	// Flag definitions.
	flags := []struct {
		name, desc, defaultVal string
		required               bool
	}{
		{"--source", "Source Redis connection string", "", true},
		{"--dest", "Destination Redis connection string", "", true},
		{"--pattern", "Key pattern to match (Redis glob pattern)", "*", false},
		{"--mode", "Migration mode", "copy", false},
		{"--conflict", "Key conflict behavior", "error", false},
		{"--batch-size", "Number of keys to process in each batch", "100", false},
		{"--concurrency", "Number of concurrent workers", "4", false},
		{"--verbose", "Enable verbose logging", "false", false},
		{"--version", "Show version information", "", false},
		{"--help", "Show this help message", "", false},
	}

	for _, flag := range flags {
		usage.WriteString("  ")
		usage.WriteString(Styles.Flag.Render(flag.name))

		// Add spacing for alignment.
		spaces := strings.Repeat(" ", int(math.Max(0, float64(15-len(flag.name)))))
		usage.WriteString(spaces)

		usage.WriteString(Styles.Description.Render(flag.desc))

		if flag.defaultVal != "" {
			usage.WriteString(" ")
			usage.WriteString(Styles.Comment.Render(fmt.Sprintf("(default: %s)", flag.defaultVal)))
		}

		if flag.required {
			usage.WriteString(" ")
			usage.WriteString(Styles.ErrorHeader.Render("REQUIRED"))
		}

		usage.WriteString("\n")
	}

	usage.WriteString("\n")
	usage.WriteString(Styles.Help.Render("Examples:"))
	usage.WriteString("\n")

	examples := []struct {
		desc    string
		command string
	}{
		{
			"Basic copy migration:",
			"redismigrate -source redis://localhost:6379/0 -dest redis://localhost:6379/1",
		},
		{
			"Move specific pattern:",
			"redismigrate -source redis://src:6379/0 -dest redis://dst:6379/0 -pattern \"user:*\" -mode move",
		},
		{
			"High throughput with custom batch size:",
			"redismigrate -source redis://src:6379/0 -dest redis://dst:6379/0 -batch-size 500 -concurrency 8",
		},
	}

	for _, example := range examples {
		usage.WriteString("  ")
		usage.WriteString(Styles.Comment.Render(example.desc))
		usage.WriteString("\n  ")
		usage.WriteString(Styles.CodeBlock.Render(example.command))
		usage.WriteString("\n\n")
	}

	// Mode options.
	usage.WriteString(Styles.Help.Render("Mode Options:"))
	usage.WriteString("\n")
	usage.WriteString("  ")
	usage.WriteString(Styles.QuotedString.Render("copy"))
	usage.WriteString("      Copy keys to destination (keeps source)")
	usage.WriteString("\n")
	usage.WriteString("  ")
	usage.WriteString(Styles.QuotedString.Render("move"))
	usage.WriteString("      Move keys to destination (removes from source)")
	usage.WriteString("\n\n")

	// Conflict options.
	usage.WriteString(Styles.Help.Render("Conflict Options:"))
	usage.WriteString("\n")
	usage.WriteString("  ")
	usage.WriteString(Styles.QuotedString.Render("error"))
	usage.WriteString("     Stop migration on key conflicts")
	usage.WriteString("\n")
	usage.WriteString("  ")
	usage.WriteString(Styles.QuotedString.Render("skip"))
	usage.WriteString("      Skip conflicting keys")
	usage.WriteString("\n")
	usage.WriteString("  ")
	usage.WriteString(Styles.QuotedString.Render("overwrite"))
	usage.WriteString(" Overwrite existing keys")
	usage.WriteString("\n")

	return usage.String()
}

// FormatVersion formats the version information for display.
func FormatVersion(version string) string {
	return fmt.Sprintf("%s %s\n",
		Styles.Program.Render("redismigrate"),
		Styles.QuotedString.Render("v"+version))
}

// FormatError formats an error for display.
func FormatError(err error) string {
	var result strings.Builder

	result.WriteString(Styles.ErrorHeader.Render("ERROR"))
	result.WriteString("\n\n")
	result.WriteString(Styles.ErrorText.Render(err.Error() + "."))
	result.WriteString("\n\n")

	result.WriteString(Styles.ErrorText.Render("Try "))
	result.WriteString(Styles.Flag.Render("--help"))
	result.WriteString(Styles.ErrorText.Render(" for usage."))
	result.WriteString("\n")

	return result.String()
}

// FormatSummary formats a migration summary combining metrics and configuration.
func FormatSummary(metrics *stats.Metrics, mode, pattern, conflict, sourceURL, destURL string) string {
	var content strings.Builder

	content.WriteString("Mode: ")
	content.WriteString(Styles.Flag.Render(mode))
	content.WriteString(" | Pattern: ")
	content.WriteString(Styles.Flag.Render(pattern))
	content.WriteString(" | Conflict: ")
	content.WriteString(Styles.Flag.Render(conflict))
	content.WriteString("\n")
	content.WriteString("Source: ")
	content.WriteString(Styles.Config.Render(sourceURL))
	content.WriteString("\n")
	content.WriteString("Destination: ")
	content.WriteString(Styles.Config.Render(destURL))
	content.WriteString("\n")

	content.WriteString("Total: ")
	content.WriteString(Styles.InfoStatus.Render(FormatCount(metrics.GetTotalKeys())))
	content.WriteString(" | Processed: ")
	content.WriteString(Styles.InfoStatus.Render(FormatCount(metrics.GetProcessedKeys())))
	content.WriteString(" | Success: ")
	content.WriteString(Styles.SuccessStatus.Render(FormatCount(metrics.GetSuccessfulKeys())))
	content.WriteString(" | Failed: ")
	content.WriteString(Styles.ErrorStatus.Render(FormatCount(metrics.GetFailedKeys())))

	switch conflict {
	case "skip":
		content.WriteString(" | Skipped: ")
		content.WriteString(Styles.InfoStatus.Render(FormatCount(metrics.GetSkippedKeys())))
	case "overwrite":
		content.WriteString(" | Overwritten: ")
		content.WriteString(Styles.InfoStatus.Render(FormatCount(metrics.GetOverwrittenKeys())))
	}

	content.WriteString("\n")
	content.WriteString("Rate: ")
	content.WriteString(Styles.InfoStatus.Render(FormatRate(metrics.GetProcessingRate())))
	content.WriteString(" | Elapsed: ")
	content.WriteString(Styles.InfoStatus.Render(FormatDuration(metrics.GetElapsed())))

	return content.String()
}
