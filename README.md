
<div align="center">
  <img src="https://vhs.charm.sh/vhs-1JVOq5s5ntrMCrMHLEyHhb.gif" alt="Made with VHS">
  <a href="https://vhs.charm.sh">
    <img src="https://stuff.charm.sh/vhs/badge.svg">
  </a>

# go-redismigrate

**Fast, reliable Redis key migration with real-time progress monitoring**

Migrate Redis keys between instances with streaming key scanning, conflict resolution, and a beautiful terminal interface.

[Features](#-features) •
[Installation](#-installation) •
[Quick Start](#-quick-start) •

</div>

---

## ✨ Features

- **🚀 High Performance**: Pipeline-based operations with streaming key scanning
- **🎯 Pattern Matching**: Support for Redis glob patterns (`user:*`, `session:*`, etc.)
- **🔄 Migration Modes**: Copy or move keys between Redis instances
- **⚡ Conflict Resolution**: Error, skip, or overwrite existing keys
- **📊 Real-time Progress**: Beautiful terminal UI with live metrics

## 📦 Installation

### Using Go Install

```bash
go install github.com/pucke-dev/go-redismigrate@latest
```

### From Source

```bash
git clone https://github.com/pucke-dev/go-redismigrate.git
cd go-redismigrate
go build -o redismigrate
```

## 🚀 Quick Start

### Basic Migration

```bash
# Copy all keys from db0 to db1
redismigrate --source redis://localhost:6379/0 --dest redis://localhost:6379/1

# Move user sessions to different Redis instance
redismigrate \
  --source redis://localhost:6379/0 \
  --dest redis://remote:6379/1 \
  --pattern "session:*" \
  --mode move
```

### With Authentication

```bash
# Migrate with Redis AUTH
redismigrate \
  --source redis://user:pass@source.redis.com:6379/0 \
  --dest redis://user:pass@dest.redis.com:6379/0 \
  --pattern "app:*"
```

## 📋 Usage

```
Usage: redismigrate [options]

Options:
  --source       Source Redis connection string  REQUIRED 
  --dest         Destination Redis connection string  REQUIRED 
  --pattern      Key pattern to match (Redis glob pattern) (default: *)
  --mode         Migration mode (default: copy)
  --conflict     Key conflict behavior (default: error)
  --batch-size   Number of keys to process in each batch (default: 100)
  --concurrency  Number of concurrent workers (default: 4)
  --verbose      Enable verbose logging (default: false)
  --version      Show version information
  --help         Show this help message

Examples:
  Basic copy migration:
   redismigrate -source redis://localhost:6379/0 -dest redis://localhost:6379/1 

  Move specific pattern:
   redismigrate -source redis://src:6379/0 -dest redis://dst:6379/0 -pattern "user:*" -mode move 

  High throughput with custom batch size:
   redismigrate -source redis://src:6379/0 -dest redis://dst:6379/0 -batch-size 500 -concurrency 8 

Mode Options:
  copy      Copy keys to destination (keeps source)
  move      Move keys to destination (removes from source)

Conflict Options:
  error     Stop migration on key conflicts
  skip      Skip conflicting keys
  overwrite Overwrite existing keys
```

## 📊 Migration Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| `copy` | Copy keys to destination, keep source | Data replication, backup |
| `move` | Move keys to destination, delete from source | Database migration, cleanup |


## ⚔️ Conflict Resolution

| Behavior | Description | Use Case |
|----------|-------------|----------|
| `error` | Stop migration on first conflict | Safe migrations, data integrity |
| `skip` | Skip existing keys, continue migration | Incremental updates |
| `overwrite` | Replace existing keys with source data | Data synchronization |



## Project Structure

```
├── internal/
│   ├── migrate/          # Core migration logic
│   ├── redis/            # Redis client with pipelining
│   ├── stats/            # Real-time metrics tracking
│   └── tui/              # Terminal user interface
├── scripts/              # Development utilities
└── main.go              # CLI entry point
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**⭐ Star this repo if you find it useful! ⭐**

</div>
