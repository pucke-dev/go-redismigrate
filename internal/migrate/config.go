package migrate

import (
	"errors"
	"fmt"
)

// Mode defines the migration mode.
type Mode int

const (
	// CopyMode copies keys from source to destination without deleting them from source.
	CopyMode Mode = iota
	// MoveMode moves keys from source to destination, deleting them from source after copying.
	MoveMode
)

// ConflictBehavior defines how to handle key conflicts in the destination.
type ConflictBehavior int

const (
	// ErrorOnConflict fails if a key already exists.
	ErrorOnConflict ConflictBehavior = iota
	// SkipOnConflict skips existing keys.
	SkipOnConflict
	// OverwriteOnConflict replaces existing keys with the new values.
	OverwriteOnConflict
)

type Config struct {
	// SourceURL is the Redis connection string for the source database.
	SourceURL string

	// DestURL is the Redis connection string for the destination database.
	DestURL string

	// Pattern is the Redis key pattern to match for migration.
	Pattern string

	// Mode specifies the migration mode. See [Mode] for details.
	Mode Mode

	// Conflict defines how to handle key conflicts in the destination. See [ConflictBehavior] for details.
	Conflict ConflictBehavior

	// BatchSize is the number of keys to process in each batch.
	BatchSize int

	// Concurrency is the number of concurrent workers to use for migration.
	Concurrency int

	// Verbose enables detailed logging during migration.
	Verbose bool
}

// Validate checks if the migration configuration is valid.
func (c *Config) Validate() error {
	errs := []error{}

	if c.SourceURL == "" {
		errs = append(errs, errors.New("no source URL provided"))
	}

	if c.DestURL == "" {
		errs = append(errs, errors.New("no destination URL provided"))
	}

	if c.Mode != CopyMode && c.Mode != MoveMode {
		errs = append(errs, fmt.Errorf("invalid mode: %d (must be CopyMode or MoveMode)", c.Mode))
	}

	if c.Conflict != ErrorOnConflict && c.Conflict != SkipOnConflict && c.Conflict != OverwriteOnConflict {
		errs = append(errs, fmt.Errorf("invalid conflict behavior: %d (must be ErrorOnConflict, SkipOnConflict, or OverwriteOnConflict)", c.Conflict))
	}

	return errors.Join(errs...)
}

// String returns the string representation of the mode.
func (m Mode) String() string {
	switch m {
	case CopyMode:
		return "copy"
	case MoveMode:
		return "move"
	default:
		return "unknown"
	}
}

// String returns the string representation of the conflict behavior.
func (c ConflictBehavior) String() string {
	switch c {
	case ErrorOnConflict:
		return "error"
	case SkipOnConflict:
		return "skip"
	case OverwriteOnConflict:
		return "overwrite"
	default:
		return "unknown"
	}
}

// ParseMode parses a string into a Mode.
func ParseMode(s string) (Mode, error) {
	switch s {
	case "copy":
		return CopyMode, nil
	case "move":
		return MoveMode, nil
	default:
		return CopyMode, fmt.Errorf("invalid mode: %s (must be 'copy' or 'move')", s)
	}
}

// ParseConflictBehavior parses a string into a ConflictBehavior.
func ParseConflictBehavior(s string) (ConflictBehavior, error) {
	switch s {
	case "error":
		return ErrorOnConflict, nil
	case "skip":
		return SkipOnConflict, nil
	case "overwrite":
		return OverwriteOnConflict, nil
	default:
		return ErrorOnConflict, fmt.Errorf("invalid conflict behavior: %s (must be 'error', 'skip', or 'overwrite')", s)
	}
}

