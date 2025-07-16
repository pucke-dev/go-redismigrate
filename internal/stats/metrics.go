package stats

import (
	"sync/atomic"
	"time"
)

// Metrics handles statistics tracking and performance calculations.
type Metrics struct {
	// totalKeys tracks the total number of keys in the source db matching the provided pattern.
	totalKeys atomic.Int64

	// processedKeys tracks the number of keys that have been processed.
	processedKeys atomic.Int64

	// successKeys tracks the number of keys that were successfully processed.
	successKeys atomic.Int64

	// failedKeys tracks the number of keys that failed processing.
	failedKeys atomic.Int64

	// skippedKeys tracks the number of keys that were skipped (e.g. already exists in destination).
	skippedKeys atomic.Int64

	// overwrittenKeys tracks the number of keys that were overwritten in the destination.
	overwrittenKeys atomic.Int64

	// startTime records when the tracking started.
	startTime time.Time
}

func NewMetrics() *Metrics {
	return &Metrics{
		startTime: time.Now(),
	}
}

func (m *Metrics) SetTotal(total int64) {
	m.totalKeys.Store(total)
}

func (m *Metrics) AddProcessed(count int64) {
	m.processedKeys.Add(count)
}

func (m *Metrics) AddSuccess(count int64) {
	m.successKeys.Add(count)
}

func (m *Metrics) AddFailed(count int64) {
	m.failedKeys.Add(count)
}

func (m *Metrics) AddSkipped(count int64) {
	m.skippedKeys.Add(count)
}

func (m *Metrics) AddOverwritten(count int64) {
	m.overwrittenKeys.Add(count)
}

func (m *Metrics) GetStartTime() time.Time {
	return m.startTime
}

func (m *Metrics) SetStartTime(t time.Time) {
	m.startTime = t
}


// GetProgress calculates completion progress as a percentage (0.0 to 1.0).
func (m *Metrics) GetProgress() float64 {
	total, processed := m.totalKeys.Load(), m.processedKeys.Load()
	if total == 0 {
		return 0.0
	}
	return float64(processed) / float64(total)
}

// GetProcessingRate calculates keys processed per second.
func (m *Metrics) GetProcessingRate() float64 {
	elapsed := m.GetElapsed().Seconds()
	if elapsed == 0 {
		return 0.0
	}
	return float64(m.processedKeys.Load()) / elapsed
}

// GetElapsed returns total elapsed time since tracking started.
func (m *Metrics) GetElapsed() time.Duration {
	return time.Since(m.startTime)
}

// GetETA calculates estimated time to completion.
func (m *Metrics) GetETA() time.Duration {
	total, processed := m.totalKeys.Load(), m.processedKeys.Load()
	rate := m.GetProcessingRate()

	if rate == 0 || total == 0 {
		return 0
	}

	remaining := total - processed
	if remaining <= 0 {
		return 0
	}

	etaSeconds := float64(remaining) / rate
	return time.Duration(etaSeconds * float64(time.Second))
}

// GetSuccessRate calculates success rate as percentage of processed keys.
func (m *Metrics) GetSuccessRate() float64 {
	processed, success := m.processedKeys.Load(), m.successKeys.Load()
	if processed == 0 {
		return 0.0
	}
	return float64(success) / float64(processed)
}

// GetFailureRate calculates failure rate as percentage of processed keys.
func (m *Metrics) GetFailureRate() float64 {
	processed, failed := m.processedKeys.Load(), m.failedKeys.Load()
	if processed == 0 {
		return 0.0
	}
	return float64(failed) / float64(processed)
}

func (m *Metrics) GetOverwrittenKeys() int64 {
	return m.overwrittenKeys.Load()
}

func (m *Metrics) GetSkippedKeys() int64 {
	return m.skippedKeys.Load()
}

func (m *Metrics) GetFailedKeys() int64 {
	return m.failedKeys.Load()
}

func (m *Metrics) GetSuccessfulKeys() int64 {
	return m.successKeys.Load()
}

func (m *Metrics) GetProcessedKeys() int64 {
	return m.processedKeys.Load()
}

func (m *Metrics) GetTotalKeys() int64 {
	return m.totalKeys.Load()
}

