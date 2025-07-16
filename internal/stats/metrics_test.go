package stats

import (
	"sync"
	"testing"
	"time"

	"testing/synctest"
)

func TestMetrics_SetTotal(t *testing.T) {
	tests := []struct {
		name  string
		total int64
	}{
		{"zero", 0},
		{"positive", 100},
		{"large", 1000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := NewMetrics()
			metrics.SetTotal(tt.total)

			if got := metrics.GetTotalKeys(); got != tt.total {
				t.Errorf("GetTotalKeys() = %d, want %d", got, tt.total)
			}
		})
	}
}

func TestMetrics_AddCounters(t *testing.T) {
	tests := []struct {
		name     string
		addFunc  func(*Metrics, int64)
		getFunc  func(*Metrics) int64
		values   []int64
		expected int64
	}{
		{
			name:     "AddProcessed",
			addFunc:  (*Metrics).AddProcessed,
			getFunc:  (*Metrics).GetProcessedKeys,
			values:   []int64{1, 5, 10},
			expected: 16,
		},
		{
			name:     "AddSuccess",
			addFunc:  (*Metrics).AddSuccess,
			getFunc:  (*Metrics).GetSuccessfulKeys,
			values:   []int64{2, 3, 7},
			expected: 12,
		},
		{
			name:     "AddFailed",
			addFunc:  (*Metrics).AddFailed,
			getFunc:  (*Metrics).GetFailedKeys,
			values:   []int64{1, 1, 1},
			expected: 3,
		},
		{
			name:     "AddSkipped",
			addFunc:  (*Metrics).AddSkipped,
			getFunc:  (*Metrics).GetSkippedKeys,
			values:   []int64{4, 6},
			expected: 10,
		},
		{
			name:     "AddOverwritten",
			addFunc:  (*Metrics).AddOverwritten,
			getFunc:  (*Metrics).GetOverwrittenKeys,
			values:   []int64{8, 2},
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := NewMetrics()

			for _, value := range tt.values {
				tt.addFunc(metrics, value)
			}

			if got := tt.getFunc(metrics); got != tt.expected {
				t.Errorf("Counter value = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestMetrics_GetProgress(t *testing.T) {
	tests := []struct {
		name      string
		total     int64
		processed int64
		expected  float64
	}{
		{"zero total", 0, 0, 0.0},
		{"zero total with processed", 0, 5, 0.0},
		{"half complete", 100, 50, 0.5},
		{"complete", 100, 100, 1.0},
		{"over complete", 100, 150, 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := NewMetrics()
			metrics.SetTotal(tt.total)
			metrics.AddProcessed(tt.processed)

			if got := metrics.GetProgress(); got != tt.expected {
				t.Errorf("GetProgress() = %f, want %f", got, tt.expected)
			}
		})
	}
}

func TestMetrics_GetProcessingRate(t *testing.T) {
	synctest.Run(func() {
		metrics := NewMetrics()
		rate := metrics.GetProcessingRate()
		if rate != 0.0 {
			t.Errorf("GetProcessingRate() = %f, want 0.0", rate)
		}

		metrics.AddProcessed(100)
		time.Sleep(time.Second)

		rate = metrics.GetProcessingRate()
		if rate != float64(100.0) {
			t.Errorf("GetProcessingRate() after 1 second = %f, want 100.0", rate)
		}

		time.Sleep(time.Second)

		rate = metrics.GetProcessingRate()
		if rate != float64(50.0) {
			t.Errorf("GetProcessingRate() after 2 seconds = %f, want 50.0", rate)
		}
	})
}

func TestMetrics_GetETA(t *testing.T) {
	tests := []struct {
		name      string
		total     int64
		processed int64
		eta       time.Duration
	}{
		{"zero total", 0, 0, 0},
		{"slow rate", 100, 1, 1*time.Minute + 39*time.Second},
		{"fast rate", 100, 80, 250 * time.Millisecond},
		{"over complete", 100, 150, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Run(func() {
				metrics := NewMetrics()
				metrics.SetTotal(tt.total)

				time.Sleep(time.Second)
				metrics.AddProcessed(tt.processed)

				if eta := metrics.GetETA(); eta != tt.eta {
					t.Errorf("GetETA() = %v, want %v", eta, tt.eta)
				}
			})
		})
	}
}

func TestMetrics_GetSuccessRate(t *testing.T) {
	tests := []struct {
		name      string
		processed int64
		success   int64
		expected  float64
	}{
		{"zero processed", 0, 0, 0.0},
		{"zero processed with success", 0, 5, 0.0},
		{"full success", 100, 100, 1.0},
		{"half success", 100, 50, 0.5},
		{"no success", 100, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := NewMetrics()
			metrics.AddProcessed(tt.processed)
			metrics.AddSuccess(tt.success)

			if got := metrics.GetSuccessRate(); got != tt.expected {
				t.Errorf("GetSuccessRate() = %f, want %f", got, tt.expected)
			}
		})
	}
}

func TestMetrics_GetFailureRate(t *testing.T) {
	tests := []struct {
		name      string
		processed int64
		failed    int64
		expected  float64
	}{
		{"zero processed", 0, 0, 0.0},
		{"zero processed with failures", 0, 5, 0.0},
		{"all failed", 100, 100, 1.0},
		{"half failed", 100, 50, 0.5},
		{"no failures", 100, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := NewMetrics()
			metrics.AddProcessed(tt.processed)
			metrics.AddFailed(tt.failed)

			if got := metrics.GetFailureRate(); got != tt.expected {
				t.Errorf("GetFailureRate() = %f, want %f", got, tt.expected)
			}
		})
	}
}

func TestMetrics_GetElapsed(t *testing.T) {
	synctest.Run(func() {
		metrics := NewMetrics()
		time.Sleep(time.Second)

		if elapsed := metrics.GetElapsed(); elapsed != time.Second {
			t.Errorf("GetElapsed() = %v, want 1s", elapsed)
		}
	})
}

func TestMetrics_ThreadSafety(t *testing.T) {
	metrics := NewMetrics()
	const goroutines = 100
	const increments = 100

	var wg sync.WaitGroup

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < increments; j++ {
				metrics.AddProcessed(1)
				metrics.AddSuccess(1)
				metrics.AddFailed(1)
				metrics.AddSkipped(1)
				metrics.AddOverwritten(1)
			}
		}()
	}

	wg.Wait()

	expected := int64(goroutines * increments)

	if got := metrics.GetProcessedKeys(); got != expected {
		t.Errorf("GetProcessedKeys() = %d, want %d", got, expected)
	}

	if got := metrics.GetSuccessfulKeys(); got != expected {
		t.Errorf("GetSuccessfulKeys() = %d, want %d", got, expected)
	}

	if got := metrics.GetFailedKeys(); got != expected {
		t.Errorf("GetFailedKeys() = %d, want %d", got, expected)
	}

	if got := metrics.GetSkippedKeys(); got != expected {
		t.Errorf("GetSkippedKeys() = %d, want %d", got, expected)
	}

	if got := metrics.GetOverwrittenKeys(); got != expected {
		t.Errorf("GetOverwrittenKeys() = %d, want %d", got, expected)
	}
}

