package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/pucke-dev/go-redismigrate/internal/migrate"
	"github.com/pucke-dev/go-redismigrate/internal/redis"
	"github.com/pucke-dev/go-redismigrate/internal/stats"
	"github.com/pucke-dev/go-redismigrate/internal/tui"
)

func main() {
	sourceURL := flag.String("source", "", "Source Redis connection string (redis://user:pass@host:port/db)")
	destURL := flag.String("dest", "", "Destination Redis connection string (redis://user:pass@host:port/db)")
	pattern := flag.String("pattern", "*", "Key pattern to match (Redis glob pattern)")
	mode := flag.String("mode", "copy", "Migration mode: copy or move")
	conflict := flag.String("conflict", "error", "Key conflict behavior: error, skip, or overwrite")
	batchSize := flag.Int("batch-size", 100, "Number of keys to process in each batch")
	concurrency := flag.Int("concurrency", 4, "Number of concurrent workers")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	version := flag.Bool("version", false, "Show version information")
	help := flag.Bool("help", false, "Show help message")

	flag.Usage = func() {
		fmt.Print(tui.FormatUsage())
	}

	flag.Parse()

	if *version {
		fmt.Print(tui.FormatVersion("0.0.0"))
		return
	}

	if *help {
		fmt.Print(tui.FormatUsage())
		return
	}

	parsedMode, err := migrate.ParseMode(*mode)
	if err != nil {
		fmt.Fprint(os.Stderr, tui.FormatError(err))
		os.Exit(1)
	}

	parsedConflict, err := migrate.ParseConflictBehavior(*conflict)
	if err != nil {
		fmt.Fprint(os.Stderr, tui.FormatError(err))
		os.Exit(1)
	}

	config := migrate.Config{
		SourceURL:   *sourceURL,
		DestURL:     *destURL,
		Pattern:     *pattern,
		Mode:        parsedMode,
		Conflict:    parsedConflict,
		BatchSize:   *batchSize,
		Concurrency: *concurrency,
		Verbose:     *verbose,
	}

	if err = config.Validate(); err != nil {
		fmt.Fprint(os.Stderr, tui.FormatError(err))
		os.Exit(1)
	}

	sourceClient, err := redis.NewClient(config.SourceURL)
	if err != nil {
		fmt.Fprint(os.Stderr, tui.FormatError(fmt.Errorf("failed to connect to source Redis: %w", err)))
		os.Exit(1)
	}
	defer sourceClient.Close()

	destClient, err := redis.NewClient(config.DestURL)
	if err != nil {
		fmt.Fprint(os.Stderr, tui.FormatError(fmt.Errorf("failed to connect to destination Redis: %w", err)))
		os.Exit(1)
	}
	defer destClient.Close()

	metrics := stats.NewMetrics()
	migrator := migrate.NewMigrator(sourceClient, destClient, config, metrics)

	model := tui.NewModel(migrator, metrics)
	program := tea.NewProgram(model, tea.WithAltScreen())

	go func() {
		ctx := context.Background()
		err := migrator.Migrate(ctx)
		if err != nil {
			program.Send(tui.ErrorCmd(err))
			return
		}

		program.Send(tui.DoneMsg{})
	}()

	if _, err := program.Run(); err != nil {
		fmt.Fprint(os.Stderr, tui.FormatError(err))
		os.Exit(1)
	}

	fmt.Fprint(os.Stderr, tui.FormatSummary(
		metrics,
		config.Mode.String(),
		config.Pattern,
		config.Conflict.String(),
		config.SourceURL,
		config.DestURL,
	))
}
