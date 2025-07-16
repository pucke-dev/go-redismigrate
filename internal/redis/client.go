package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/pucke-dev/go-redismigrate/internal/migrate"
)

type Client struct {
	client *redis.Client
}

func NewClient(connStr string) (*Client, error) {
	opts, err := redis.ParseURL(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{
		client: client,
	}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

// ScanKeys streams keys matching a pattern to a channel in batches.
func (c *Client) ScanKeys(ctx context.Context, pattern string, batchSize int, keysChan chan<- []string) error {
	var cursor uint64

	for {
		keys, newCursor, err := c.client.Scan(ctx, cursor, pattern, int64(batchSize)).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			select {
			case keysChan <- keys:
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

func (c *Client) CountKeys(ctx context.Context, pattern string, batchSize int) (int64, error) {
	var count int64
	var cursor uint64

	for {
		keys, newCursor, err := c.client.Scan(ctx, cursor, pattern, int64(batchSize)).Result()
		if err != nil {
			return 0, err
		}

		count += int64(len(keys))
		cursor = newCursor

		if cursor == 0 {
			break
		}
	}

	return count, nil
}

func (c *Client) DumpKeys(ctx context.Context, keys []string) ([]migrate.KeyData, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	pipe := c.client.Pipeline()

	for _, key := range keys {
		pipe.Dump(ctx, key)
		pipe.PTTL(ctx, key)
	}

	results, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("failed to dump keys: %w", err)
	}

	var keyData []migrate.KeyData
	for i, key := range keys {
		dumpCmd := results[i*2].(*redis.StringCmd)
		ttlCmd := results[i*2+1].(*redis.DurationCmd)

		// Skip keys where DUMP or PTTL failed (e.g., key doesn't exist)
		if dumpCmd.Err() != nil || ttlCmd.Err() != nil {
			continue
		}

		keyData = append(keyData, migrate.KeyData{
			Key:  key,
			Data: dumpCmd.Val(),
			TTL:  ttlCmd.Val(),
		})
	}

	return keyData, nil
}

func (c *Client) RestoreKeys(ctx context.Context, data []migrate.KeyData, behavior migrate.ConflictBehavior) ([]string, error) {
	if len(data) == 0 {
		return nil, nil
	}

	pipe := c.client.Pipeline()
	var keysToRestore []migrate.KeyData

	for _, info := range data {
		switch behavior {
		case migrate.SkipOnConflict:
			if c.client.Exists(ctx, info.Key).Val() > 0 {
				continue
			}
			pipe.Restore(ctx, info.Key, info.TTL, info.Data)

		case migrate.OverwriteOnConflict:
			pipe.RestoreReplace(ctx, info.Key, info.TTL, info.Data)

		case migrate.ErrorOnConflict:
			pipe.Restore(ctx, info.Key, info.TTL, info.Data)
		}

		keysToRestore = append(keysToRestore, info)
	}

	results, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to restore keys: %w", err)
	}

	var successfulKeys []string
	for i, result := range results {
		if result.Err() != nil {
			continue
		}
		successfulKeys = append(successfulKeys, keysToRestore[i].Key)
	}

	return successfulKeys, nil
}

func (c *Client) DeleteKeys(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	return c.client.Del(ctx, keys...).Err()
}
