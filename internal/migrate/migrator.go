package migrate

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/pucke-dev/go-redismigrate/internal/stats"
)

// RedisClient defines the interface for Redis operations needed by the migrator.
type RedisClient interface {
	// ScanKeys streams keys matching a pattern to a channel in batches.
	ScanKeys(ctx context.Context, pattern string, batchSize int, keysChan chan<- []string) error

	// CountKeys counts keys matching a pattern.
	CountKeys(ctx context.Context, pattern string, batchSize int) (int64, error)

	// DumpKeys dumps multiple keys and returns their data and TTL.
	DumpKeys(ctx context.Context, keys []string) ([]KeyData, error)

	// RestoreKeys restores multiple keys with conflict handling.
	RestoreKeys(ctx context.Context, data []KeyData, behavior ConflictBehavior) ([]string, error)

	// DeleteKeys deletes multiple keys.
	DeleteKeys(ctx context.Context, keys []string) error

	// Close closes the client connection.
	Close() error
}

// KeyData represents a Redis key with its data and TTL.
type KeyData struct {
	Key  string
	Data string
	TTL  time.Duration
}

type Migrator struct {
	// source is the redis client for the source database.
	source RedisClient

	// dest is the redis client for the destination database.
	dest RedisClient

	// config contains the migration configuration.
	config Config

	// metrics is used to track migration progress and statistics.
	metrics *stats.Metrics

	errors []error
	mu     sync.Mutex
}

func NewMigrator(source, dest RedisClient, config Config, metrics *stats.Metrics) *Migrator {
	return &Migrator{
		source:  source,
		dest:    dest,
		config:  config,
		metrics: metrics,
		errors:  make([]error, 0),
	}
}

func (m *Migrator) GetConfig() Config {
	return m.config
}

func (m *Migrator) addError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors = append(m.errors, err)
}

func (m *Migrator) GetErrors() []error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return slices.Clone(m.errors)
}

func (m *Migrator) Migrate(ctx context.Context) error {
	totalKeys, err := m.source.CountKeys(ctx, m.config.Pattern, m.config.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to count keys: %w", err)
	}

	m.metrics.SetTotal(totalKeys)

	if totalKeys == 0 {
		return nil
	}

	keysChan := make(chan []string, m.config.Concurrency*2)
	errorsChan := make(chan error, m.config.Concurrency)

	var wg sync.WaitGroup

	wg.Add(m.config.Concurrency)
	for range m.config.Concurrency {
		go func() {
			defer wg.Done()
			m.worker(ctx, keysChan, errorsChan)
		}()
	}

	go func() {
		defer close(keysChan)
		err := m.source.ScanKeys(ctx, m.config.Pattern, m.config.BatchSize, keysChan)
		if err != nil {
			errorsChan <- fmt.Errorf("failed to scan keys: %w", err)
		}
	}()

	// Wait for workers to finish.
	go func() {
		wg.Wait()
		close(errorsChan)
	}()

	for err := range errorsChan {
		if err != nil {
			m.addError(err)
		}
	}

	return errors.Join(m.GetErrors()...)
}

func (m *Migrator) worker(ctx context.Context, keysChan <-chan []string, errorsChan chan<- error) {
	for {
		select {
		case keys, ok := <-keysChan:
			if !ok {
				return
			}
			m.processBatch(ctx, keys, errorsChan)
		case <-ctx.Done():
			return
		}
	}
}

func (m *Migrator) processBatch(ctx context.Context, keys []string, errorsChan chan<- error) {
	keyData, err := m.source.DumpKeys(ctx, keys)
	if err != nil {
		errorsChan <- fmt.Errorf("failed to dump keys: %w", err)
		m.metrics.AddProcessed(int64(len(keys)))
		m.metrics.AddFailed(int64(len(keys)))
		return
	}

	successfulKeys, err := m.dest.RestoreKeys(ctx, keyData, m.config.Conflict)
	if err != nil {
		errorsChan <- fmt.Errorf("failed to restore keys: %w", err)
		m.metrics.AddProcessed(int64(len(keyData)))
		m.metrics.AddFailed(int64(len(keyData)))
		return
	}

	m.updateMetricsForBatch(keyData, successfulKeys)

	// Delete from source if move mode.
	if m.config.Mode == MoveMode && len(successfulKeys) > 0 {
		if err := m.source.DeleteKeys(ctx, successfulKeys); err != nil {
			errorsChan <- fmt.Errorf("failed to delete keys from source: %w", err)
		}
	}
}

func (m *Migrator) updateMetricsForBatch(keyData []KeyData, successfulKeys []string) {
	successCount := int64(len(successfulKeys))
	totalCount := int64(len(keyData))

	m.metrics.AddProcessed(totalCount)

	if m.config.Conflict == OverwriteOnConflict {
		m.metrics.AddOverwritten(successCount)
	} else {
		m.metrics.AddSuccess(successCount)
	}

	failedCount := totalCount - successCount
	if failedCount > 0 {
		m.metrics.AddFailed(failedCount)
	}
}
