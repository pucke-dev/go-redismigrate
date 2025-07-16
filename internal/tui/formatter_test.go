package tui

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"testing/synctest"
	"time"

	"github.com/pucke-dev/go-redismigrate/internal/stats"
)

var update = flag.Bool("update", false, "update golden files")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func goldenPath(name string) string {
	return filepath.Join("testdata", name+".golden")
}

func updateGolden(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create testdata directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to update golden file %s: %v", path, err)
	}
}

func loadGolden(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", path, err)
	}
	return string(content)
}

func assertGolden(t *testing.T, name, got string) {
	t.Helper()
	path := goldenPath(name)
	
	if *update {
		updateGolden(t, path, got)
		return
	}
	
	expected := loadGolden(t, path)
	if got != expected {
		t.Errorf("output mismatch for %s:\nGot:\n%s\nWant:\n%s", name, got, expected)
	}
}

func TestFormatUsage(t *testing.T) {
	got := FormatUsage()
	assertGolden(t, "format_usage", got)
}

func TestFormatVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{"basic_version", "1.0.0"},
		{"dev_version", "dev"},
		{"semver_with_prerelease", "2.1.0-alpha.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatVersion(tt.version)
			assertGolden(t, "format_version_"+tt.name, got)
		})
	}
}

func TestFormatError(t *testing.T) {
	tests := []struct {
		name string
		err  string
	}{
		{"simple_error", "connection failed"},
		{"validation_error", "invalid configuration: missing source URL"},
		{"long_error", "failed to connect to Redis server at localhost:6379: connection timeout after 5 seconds"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatError(errors.New(tt.err))
			assertGolden(t, "format_error_"+tt.name, got)
		})
	}
}

func TestFormatSummary(t *testing.T) {
	tests := []struct {
		name      string
		mode      string
		pattern   string
		conflict  string
		sourceURL string
		destURL   string
	}{
		{
			"copy_mode_error_conflict",
			"copy",
			"user:*",
			"error",
			"redis://localhost:6379/0",
			"redis://localhost:6379/1",
		},
		{
			"move_mode_skip_conflict",
			"move",
			"session:*",
			"skip",
			"redis://source.example.com:6379/2",
			"redis://dest.example.com:6379/3",
		},
		{
			"copy_mode_overwrite_conflict",
			"copy",
			"cache:*",
			"overwrite",
			"redis://10.0.1.100:6379/0",
			"redis://10.0.1.200:6379/0",
		},
		{
			"wildcard_pattern",
			"copy",
			"*",
			"error",
			"redis://redis1:6379/0",
			"redis://redis2:6379/0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Run(func() {
				// Create metrics with known values for consistent testing
				testMetrics := stats.NewMetrics()
				testMetrics.SetTotal(1000)
				testMetrics.AddProcessed(750)
				testMetrics.AddSuccess(700)
				testMetrics.AddFailed(25)
				testMetrics.AddSkipped(15)
				testMetrics.AddOverwritten(10)
				
				// Advance time by 5 minutes for consistent elapsed time
				time.Sleep(5 * time.Minute)
				
				got := FormatSummary(testMetrics, tt.mode, tt.pattern, tt.conflict, tt.sourceURL, tt.destURL)
				assertGolden(t, "format_summary_"+tt.name, got)
			})
		})
	}
}

func TestFormatCount(t *testing.T) {
	tests := []struct {
		name  string
		count int64
		want  string
	}{
		{"zero", 0, "0"},
		{"positive", 123, "123"},
		{"large", 1000000, "1000000"},
		{"negative", -50, "-50"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatCount(tt.count)
			if got != tt.want {
				t.Errorf("FormatCount(%d) = %q, want %q", tt.count, got, tt.want)
			}
		})
	}
}

func TestFormatRate(t *testing.T) {
	tests := []struct {
		name string
		rate float64
		want string
	}{
		{"zero", 0.0, "0.0 keys/sec"},
		{"decimal", 15.7, "15.7 keys/sec"},
		{"integer", 100.0, "100.0 keys/sec"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatRate(tt.rate)
			if got != tt.want {
				t.Errorf("FormatRate(%f) = %q, want %q", tt.rate, got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"zero", 0, "0s"},
		{"seconds", 5 * time.Second, "5s"},
		{"minutes", 2 * time.Minute, "2m0s"},
		{"hours", time.Hour, "1h0m0s"},
		{"subsecond_rounds_up", 500 * time.Millisecond, "1s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}