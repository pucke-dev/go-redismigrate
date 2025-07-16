package redis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/redis"

	"github.com/pucke-dev/go-redismigrate/internal/migrate"
)

type RedisClientTestSuite struct {
	suite.Suite
	redisContainer *redis.RedisContainer
	client         *Client
	ctx            context.Context
}

func (s *RedisClientTestSuite) SetupSuite() {
	s.ctx = context.Background()

	redisContainer, err := redis.Run(s.ctx, "redis:7-alpine")
	if err != nil {
		s.T().Fatalf("Failed to start Redis container: %v", err)
	}
	s.redisContainer = redisContainer

	connectionString, err := redisContainer.ConnectionString(s.ctx)
	if err != nil {
		s.T().Fatalf("Failed to get connection string: %v", err)
	}

	client, err := NewClient(connectionString)
	if err != nil {
		s.T().Fatalf("Failed to create Redis client: %v", err)
	}
	s.client = client
}

func (s *RedisClientTestSuite) TearDownSuite() {
	if s.client != nil {
		s.client.Close()
	}
	if s.redisContainer != nil {
		if err := s.redisContainer.Terminate(s.ctx); err != nil {
			s.T().Logf("Failed to terminate container: %v", err)
		}
	}
}

func (s *RedisClientTestSuite) TearDownTest() {
	err := s.client.client.FlushDB(s.ctx).Err()
	if err != nil {
		s.T().Fatalf("Failed to flush Redis database: %v", err)
	}
}

func (s *RedisClientTestSuite) setupTestData() map[string]string {
	testKeys := map[string]string{
		"user:1":    "alice",
		"user:2":    "bob",
		"session:1": "active",
		"config:db": "production",
	}

	for key, value := range testKeys {
		err := s.client.client.Set(s.ctx, key, value, 0).Err()
		if err != nil {
			s.T().Fatalf("Failed to set test data: %v", err)
		}
	}

	return testKeys
}

func (s *RedisClientTestSuite) TestNewClient_Success() {
	t := s.T()

	connectionString, err := s.redisContainer.ConnectionString(s.ctx)
	assert.NoError(t, err)

	client, err := NewClient(connectionString)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	err = client.Close()
	assert.NoError(t, err, "Failed to close Redis client")
}

func (s *RedisClientTestSuite) TestScanKeys_EmptyDatabase() {
	t := s.T()

	keysChan := make(chan []string, 10)
	var allKeys []string

	go func() {
		defer close(keysChan)
		err := s.client.ScanKeys(s.ctx, "*", 100, keysChan)
		assert.NoError(t, err)
	}()

	for keys := range keysChan {
		allKeys = append(allKeys, keys...)
	}

	assert.Empty(t, allKeys, "Expected empty database")
}

func (s *RedisClientTestSuite) TestScanKeys_WithData() {
	t := s.T()

	s.setupTestData()

	tests := []struct {
		name     string
		pattern  string
		expected int
	}{
		{"all keys", "*", 4},
		{"user keys", "user:*", 2},
		{"session keys", "session:*", 1},
		{"config keys", "config:*", 1},
		{"non-existent", "notfound:*", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keysChan := make(chan []string, 10)
			var allKeys []string

			go func() {
				defer close(keysChan)
				err := s.client.ScanKeys(s.ctx, tt.pattern, 100, keysChan)
				assert.NoError(t, err)
			}()

			for keys := range keysChan {
				allKeys = append(allKeys, keys...)
			}

			assert.Len(t, allKeys, tt.expected, "Pattern %s should match %d keys", tt.pattern, tt.expected)
		})
	}
}

func (s *RedisClientTestSuite) TestCountKeys_WithPatterns() {
	t := s.T()

	s.setupTestData()

	tests := []struct {
		name     string
		pattern  string
		expected int64
	}{
		{"all keys", "*", 4},
		{"user keys", "user:*", 2},
		{"session keys", "session:*", 1},
		{"config keys", "config:*", 1},
		{"non-existent", "notfound:*", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := s.client.CountKeys(s.ctx, tt.pattern, 100)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, count, "Pattern %s should count %d keys", tt.pattern, tt.expected)
		})
	}
}

func (s *RedisClientTestSuite) TestDumpKeys_Operations() {
	t := s.T()

	keyData, err := s.client.DumpKeys(s.ctx, []string{})
	assert.NoError(t, err)
	assert.Nil(t, keyData)

	s.setupTestData()

	// Test with only existing keys first
	keys := []string{"user:1", "user:2"}
	keyData, err = s.client.DumpKeys(s.ctx, keys)
	assert.NoError(t, err)
	assert.Len(t, keyData, 2, "Should dump 2 existing keys")

	for _, data := range keyData {
		assert.Contains(t, []string{"user:1", "user:2"}, data.Key)
		assert.NotEmpty(t, data.Data, "Dump data should not be empty")
	}

	// Test with mix of existing and non-existing keys
	keys = []string{"user:1", "user:2", "nonexistent"}
	keyData, err = s.client.DumpKeys(s.ctx, keys)
	assert.NoError(t, err)
	assert.Len(t, keyData, 2, "Should dump 2 existing keys, skip nonexistent")
}

func (s *RedisClientTestSuite) TestRestoreKeys_ConflictBehaviors() {
	t := s.T()

	// Setup: create a key that will conflict
	err := s.client.client.Set(s.ctx, "conflict:key", "original", 0).Err()
	assert.NoError(t, err)

	// Dump the existing key to get real Redis dump data
	existingData, err := s.client.DumpKeys(s.ctx, []string{"conflict:key"})
	assert.NoError(t, err)
	assert.Len(t, existingData, 1)

	testData := []migrate.KeyData{
		{
			Key:  "conflict:key",
			Data: existingData[0].Data,
			TTL:  0,
		},
		{
			Key:  "new:key",
			Data: existingData[0].Data,
			TTL:  0,
		},
	}

	tests := []struct {
		name               string
		behavior           migrate.ConflictBehavior
		expectedSuccessful []string
		expectError        bool
		setupCleanup       bool
	}{
		{
			name:               "ErrorOnConflict",
			behavior:           migrate.ErrorOnConflict,
			expectedSuccessful: nil, // Should fail entirely on conflict
			expectError:        true,
			setupCleanup:       true,
		},
		{
			name:               "SkipOnConflict",
			behavior:           migrate.SkipOnConflict,
			expectedSuccessful: []string{"new:key"}, // Should skip conflict:key
			expectError:        false,
			setupCleanup:       true,
		},
		{
			name:               "OverwriteOnConflict",
			behavior:           migrate.OverwriteOnConflict,
			expectedSuccessful: []string{"conflict:key", "new:key"}, // Should restore both
			expectError:        false,
			setupCleanup:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupCleanup {
				s.client.client.Del(s.ctx, "new:key")
			}

			successfulKeys, err := s.client.RestoreKeys(s.ctx, testData, tt.behavior)

			if tt.expectError {
				assert.Error(t, err, "Behavior %v should return error on conflict", tt.behavior)
				assert.Nil(t, successfulKeys, "Should return nil keys on error")
			} else {
				assert.NoError(t, err, "Behavior %v should not return error", tt.behavior)
				assert.ElementsMatch(t, tt.expectedSuccessful, successfulKeys,
					"Behavior %v should restore expected keys", tt.behavior)
			}
		})
	}
}

func (s *RedisClientTestSuite) TestDeleteKeys_Operations() {
	t := s.T()

	err := s.client.DeleteKeys(s.ctx, []string{})
	assert.NoError(t, err)

	testKeys := map[string]string{
		"user:1": "alice",
		"user:2": "bob",
		"user:3": "charlie",
	}

	for key, value := range testKeys {
		err = s.client.client.Set(s.ctx, key, value, 0).Err()
		assert.NoError(t, err)
	}

	keysToDelete := []string{"user:1", "user:2"}
	err = s.client.DeleteKeys(s.ctx, keysToDelete)
	assert.NoError(t, err)

	count, err := s.client.CountKeys(s.ctx, "user:*", 100)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count, "Should have 1 remaining user key")

	remaining, err := s.client.client.Get(s.ctx, "user:3").Result()
	assert.NoError(t, err)
	assert.Equal(t, "charlie", remaining)
}

func (s *RedisClientTestSuite) TestCompleteWorkflow() {
	t := s.T()

	sourceData := map[string]string{
		"app:config:db":    "production",
		"app:config:cache": "redis",
		"user:1000":        "admin",
		"user:1001":        "user",
	}

	for key, value := range sourceData {
		err := s.client.client.Set(s.ctx, key, value, 30*time.Second).Err()
		assert.NoError(t, err)
	}

	count, err := s.client.CountKeys(s.ctx, "app:config:*", 100)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	keysChan := make(chan []string, 10)
	var scannedKeys []string

	go func() {
		defer close(keysChan)
		err = s.client.ScanKeys(s.ctx, "app:config:*", 100, keysChan)
		assert.NoError(t, err)
	}()

	for keys := range keysChan {
		scannedKeys = append(scannedKeys, keys...)
	}

	assert.Len(t, scannedKeys, 2)

	keyData, err := s.client.DumpKeys(s.ctx, scannedKeys)
	assert.NoError(t, err)
	assert.Len(t, keyData, 2)

	err = s.client.DeleteKeys(s.ctx, scannedKeys)
	assert.NoError(t, err)

	count, err = s.client.CountKeys(s.ctx, "app:config:*", 100)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	successfulKeys, err := s.client.RestoreKeys(s.ctx, keyData, migrate.ErrorOnConflict)
	assert.NoError(t, err)
	assert.Len(t, successfulKeys, 2)

	count, err = s.client.CountKeys(s.ctx, "app:config:*", 100)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	for _, key := range scannedKeys {
		exists := s.client.client.Exists(s.ctx, key).Val()
		assert.Equal(t, int64(1), exists, "Key %s should exist after restore", key)

		ttl := s.client.client.TTL(s.ctx, key).Val()
		assert.Greater(t, ttl, time.Duration(0), "Key %s should have TTL preserved", key)
	}
}

func TestRedisClientTestSuite(t *testing.T) {
	suite.Run(t, new(RedisClientTestSuite))
}
