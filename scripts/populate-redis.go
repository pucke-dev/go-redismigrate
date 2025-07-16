package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <redis-url> <num-keys> [prefix]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s redis://localhost:6379 1000 test\n", os.Args[0])
		os.Exit(1)
	}

	redisURL := os.Args[1]
	numKeys, err := strconv.Atoi(os.Args[2])
	if err != nil {
		slog.Error("Invalid number of keys", "error", err)
		os.Exit(1)
	}

	prefix := "key"
	if len(os.Args) > 3 {
		prefix = os.Args[3]
	}

	// Connect to Redis
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		slog.Error("Failed to parse Redis URL", "error", err)
		os.Exit(1)
	}

	client := redis.NewClient(opt)
	defer client.Close()

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		slog.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}

	fmt.Printf("Inserting %d keys with prefix '%s' into Redis...\n", numKeys, prefix)

	// Insert keys in batches
	batchSize := 100
	for i := 0; i < numKeys; i += batchSize {
		pipe := client.Pipeline()

		end := i + batchSize
		if end > numKeys {
			end = numKeys
		}

		for j := i; j < end; j++ {
			key := fmt.Sprintf("%s:%d", prefix, j)
			value := fmt.Sprintf("value-%d", j)

			pipe.Set(ctx, key, value, 0) // No TTL
		}

		if _, err := pipe.Exec(ctx); err != nil {
			slog.Error("Failed to insert batch", "error", err)
			os.Exit(1)
		}

		if i%1000 == 0 {
			fmt.Printf("Inserted %d keys...\n", i)
		}
	}

	fmt.Printf("Successfully inserted %d keys\n", numKeys)
}
