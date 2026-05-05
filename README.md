# 🛡️ DBVault

> A production-grade CLI utility for backing up, restoring, and scheduling database operations — built in Go.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Platforms](https://img.shields.io/badge/platform-linux%20%7C%20macos%20%7C%20windows-lightgrey)]()
[![DBs Supported](https://img.shields.io/badge/databases-MySQL%20%7C%20PostgreSQL%20%7C%20MongoDB%20%7C%20SQLite-blue)]()

---

## 📌 Table of Contents

- [Overview](#-overview)
- [Architecture](#-architecture)
- [Project Structure](#-project-structure)
- [Features](#-features)
- [Supported Databases](#-supported-databases)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [CLI Reference](#-cli-reference)
- [Backup Types](#-backup-types)
- [Storage Backends](#-storage-backends)
- [Scheduling](#-scheduling)
- [Logging](#-logging)
- [Notifications](#-notifications)
- [Restore Operations](#-restore-operations)
- [Design Decisions](#-design-decisions)
- [Roadmap](#-roadmap)

---

## 🔍 Overview

**DBVault** is a single-binary CLI tool written in Go that lets you back up any supported database with one command. It supports multiple database engines, pluggable storage backends (local + cloud), automatic scheduling via cron, gzip compression, structured logging, and Slack notifications — all configured through a clean YAML file or CLI flags.

```bash
# Back up a MySQL database to S3, with gzip compression, and notify Slack
dbvault backup \
  --db mysql \
  --host localhost \
  --user root \
  --password secret \
  --name production_db \
  --type full \
  --storage s3 \
  --compress gzip \
  --notify
```

---

## 🏗️ Architecture

### High-Level System Design

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLI Layer (Cobra)                         │
│   backup │ restore │ schedule │ test-connection │ config         │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Backup Manager                              │
│  Orchestrates: Validate → Dump → Compress → Store → Log → Notify│
└────┬──────────────┬──────────────┬──────────────┬───────────────┘
     │              │              │              │
     ▼              ▼              ▼              ▼
┌─────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐
│  DB     │  │Compressor│  │ Storage  │  │  Notifier    │
│Connector│  │  (gzip)  │  │ Backend  │  │  (Slack)     │
│Interface│  └──────────┘  │Interface │  └──────────────┘
│         │                └────┬─────┘
│MySQL    │                     │
│Postgres │              ┌──────┴──────┐
│MongoDB  │              │             │
│SQLite   │           Local          AWS S3
└─────────┘            Storage     (GCS/Azure
                                    planned)
```

### Request Flow — Backup Operation

```
User CLI Input
      │
      ▼
┌─────────────────┐
│  Flag Parsing   │  --db, --host, --user, --type, --storage, ...
│  + Viper Config │  Merges flags > env vars > config file > defaults
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ DBConfig Build  │  Constructs typed config struct
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Connection Test │  Ping DB, validate credentials
│ (Pre-flight)    │  Fail fast before doing any I/O
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Backup Engine  │  Shells out to: mysqldump / pg_dump /
│  (per DBMS)     │  mongodump / file-copy (SQLite)
└────────┬────────┘
         │  raw dump stream (io.Reader)
         ▼
┌─────────────────┐
│   Compressor    │  Wraps io.Reader → gzip.Writer
│   (streaming)   │  No full-file buffering in memory
└────────┬────────┘
         │  compressed stream
         ▼
┌─────────────────┐
│ Storage Backend │  Writes to: Local disk / S3 / GCS / Azure
│  (io.Writer)    │  Generates: SHA-256 checksum
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Metadata Writer │  Saves .meta.json alongside backup:
│                 │  id, type, size, checksum, duration, path
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Structured Log  │  Appends to ~/.dbvault/logs/dbvault.log
│  (slog/JSON)    │  Fields: timestamp, level, db, status, duration
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Slack Notifier  │  POST to webhook if --notify flag set
│  (optional)     │  Payload: backup summary + download link
└─────────────────┘
```

### Restore Operation Flow

```
File Path (local or S3 key)
      │
      ▼
┌─────────────────┐
│ Metadata Loader │  Reads .meta.json → validates checksum
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Storage Fetcher │  Downloads from S3 / reads local file
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Decompressor   │  gzip.Reader wrapping the download stream
└────────┬────────┘
         │  raw SQL / archive stream
         ▼
┌─────────────────┐
│ Restore Engine  │  Pipes to: mysql / psql / mongorestore / cp
│  (per DBMS)     │  Selective restore: --tables flag filters output
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Log + Notify    │  Same pipeline as backup
└─────────────────┘
```

### Scheduler Architecture

```
dbvault schedule add --cron "0 2 * * *" --db mysql ...
         │
         ▼
┌─────────────────────────────────────────┐
│         Schedule Registry               │
│   ~/.dbvault/schedules.json             │
│   [{ id, cron, db_config, backup_opts}] │
└────────────────┬────────────────────────┘
                 │  loaded on daemon start
                 ▼
┌─────────────────────────────────────────┐
│      robfig/cron Scheduler              │
│   Registers each schedule as a cron job │
│   Runs in background goroutine          │
└────────────────┬────────────────────────┘
                 │  on trigger (e.g. 02:00 daily)
                 ▼
         Backup Manager
         (same pipeline as manual backup)
```

---

## 📁 Project Structure

```
dbvault/
│
├── main.go                          # Binary entry point
│
├── cmd/                             # Cobra CLI commands (thin layer)
│   ├── root.go                      # Root command, global flags, config init
│   ├── backup.go                    # `dbvault backup` command
│   ├── restore.go                   # `dbvault restore` command
│   ├── schedule.go                  # `dbvault schedule [add|list|remove]`
│   ├── testconn.go                  # `dbvault test-connection` command
│   └── config.go                    # `dbvault config [get|set|view]` command
│
├── internal/                        # Core business logic (not exported)
│   │
│   ├── models/
│   │   └── models.go                # Shared types: DBConfig, BackupRecord, etc.
│   │
│   ├── config/
│   │   └── config.go                # Viper config loader + AppConfig struct
│   │
│   ├── db/                          # Database connector abstraction
│   │   ├── connector.go             # Connector interface + factory function
│   │   ├── mysql.go                 # MySQL: mysqldump + mysql restore
│   │   ├── postgres.go              # PostgreSQL: pg_dump + psql restore
│   │   ├── mongodb.go               # MongoDB: mongodump + mongorestore
│   │   └── sqlite.go                # SQLite: file copy + sql import
│   │
│   ├── backup/
│   │   └── manager.go               # BackupManager: orchestrates full pipeline
│   │
│   ├── compress/
│   │   └── compress.go              # Streaming gzip compress + decompress
│   │
│   ├── storage/                     # Pluggable storage backends
│   │   ├── storage.go               # Storage interface definition
│   │   ├── local.go                 # Local filesystem backend
│   │   └── s3.go                    # AWS S3 backend (v2 SDK)
│   │
│   ├── scheduler/
│   │   └── scheduler.go             # Cron scheduler + schedule persistence
│   │
│   ├── logger/
│   │   └── logger.go                # Structured logger (slog) + file handler
│   │
│   └── notify/
│       └── slack.go                 # Slack webhook notification client
│
├── .dbvault.example.yaml            # Annotated sample config file
├── Makefile                         # Build, test, lint, install targets
└── README.md
```

---

## ✨ Features

| Feature | Details |
|---|---|
| **Multi-DBMS** | MySQL, PostgreSQL, MongoDB, SQLite |
| **Backup Types** | Full, Incremental (MySQL/Mongo), Differential (Postgres) |
| **Compression** | Streaming gzip — no large memory footprint |
| **Storage** | Local disk, AWS S3 (GCS + Azure on roadmap) |
| **Scheduling** | Cron-based scheduler with persistent registry |
| **Logging** | Structured JSON logs with rotation support |
| **Notifications** | Slack webhook on backup success or failure |
| **Restore** | Full restore + selective table/collection restore |
| **Integrity** | SHA-256 checksum verified before every restore |
| **Config** | YAML file, environment variables, or inline flags |
| **Cross-platform** | Linux, macOS, Windows (amd64 + arm64) |

---

## 🗄️ Supported Databases

| DBMS | Backup Tool | Restore Tool | Full | Incremental | Differential |
|---|---|---|---|---|---|
| MySQL 5.7+ | `mysqldump` | `mysql` | ✅ | ✅ (binlog) | ❌ |
| PostgreSQL 12+ | `pg_dump` | `psql` | ✅ | ❌ | ✅ (WAL) |
| MongoDB 4.4+ | `mongodump` | `mongorestore` | ✅ | ✅ (oplog) | ❌ |
| SQLite 3 | file copy | file replace | ✅ | ❌ | ❌ |

> **Note:** Incremental and Differential backups require additional database configuration.
> See [Backup Types](#-backup-types) for prerequisites.

---

## 📦 Installation

### Option 1: Build from Source

```bash
git clone https://github.com/dbvault/dbvault.git
cd dbvault

# Download dependencies
go mod tidy

# Build binary
make build

# Install to $GOPATH/bin
make install
```

### Option 2: Download Pre-built Binary

```bash
# Linux (amd64)
curl -L https://github.com/dbvault/dbvault/releases/latest/download/dbvault-linux-amd64 -o dbvault
chmod +x dbvault && sudo mv dbvault /usr/local/bin/

# macOS (arm64 / Apple Silicon)
curl -L https://github.com/dbvault/dbvault/releases/latest/download/dbvault-darwin-arm64 -o dbvault
chmod +x dbvault && sudo mv dbvault /usr/local/bin/
```

### Prerequisites

DBVault shells out to native database tools for backup/restore. Ensure these are in your `$PATH`:

| Database | Required Binaries |
|---|---|
| MySQL | `mysqldump`, `mysql` |
| PostgreSQL | `pg_dump`, `psql` |
| MongoDB | `mongodump`, `mongorestore` |
| SQLite | none (handled natively) |

---

## ⚙️ Configuration

DBVault supports three ways to pass configuration, merged in priority order:

```
CLI flags  >  Environment Variables  >  Config File  >  Defaults
```

### Config File Location

```
~/.dbvault/config.yaml        (default, auto-created on first run)
./dbvault.yaml                (project-level override)
Custom path via --config flag
```

### Sample Config: `.dbvault.example.yaml`

```yaml
# Database connection defaults
database:
  type: mysql           # mysql | postgres | mongodb | sqlite
  host: localhost
  port: 3306
  username: root
  password: ""          # Prefer DBVAULT_DB_PASSWORD env var
  name: ""

# Backup defaults
backup:
  type: full            # full | incremental | differential
  compression: gzip     # gzip | none
  output_dir: ~/.dbvault/backups

# Storage backend
storage:
  type: local           # local | s3 | gcs | azure

  local:
    path: ~/.dbvault/backups

  s3:
    bucket: my-db-backups
    region: ap-south-1
    access_key: ""      # Prefer AWS_ACCESS_KEY_ID env var
    secret_key: ""      # Prefer AWS_SECRET_ACCESS_KEY env var
    prefix: backups/    # Optional key prefix

# Notifications
notifications:
  slack:
    enabled: false
    webhook_url: ""     # Prefer DBVAULT_SLACK_WEBHOOK env var

# Logging
logging:
  level: info           # debug | info | warn | error
  format: json          # json | text
  file: ~/.dbvault/logs/dbvault.log
  max_size_mb: 100
  max_backups: 7
```

### Environment Variables

```bash
DBVAULT_DB_PASSWORD=secret
DBVAULT_SLACK_WEBHOOK=https://hooks.slack.com/...
AWS_ACCESS_KEY_ID=AKIA...
AWS_SECRET_ACCESS_KEY=...
AWS_REGION=ap-south-1
```

---

## 📟 CLI Reference

```
dbvault [command] [flags]

Available Commands:
  backup           Create a database backup
  restore          Restore a database from a backup file
  schedule         Manage automated backup schedules
  test-connection  Validate database connectivity
  config           View or modify configuration
  help             Show help for any command

Global Flags:
  --config string   Config file path (default: ~/.dbvault/config.yaml)
  --log-level       Log verbosity: debug|info|warn|error (default: info)
  -h, --help        Help for any command
  -v, --version     Print version information
```

---

### `backup` — Create a Backup

```bash
dbvault backup [flags]

Flags:
  -d, --db string           Database type: mysql|postgres|mongodb|sqlite (required)
      --host string         Database host (default: localhost)
  -p, --port int            Database port (uses DB default if omitted)
  -u, --user string         Database username
  -P, --password string     Database password (prefer env var)
  -n, --name string         Database name to back up
  -t, --type string         Backup type: full|incremental|differential (default: full)
  -s, --storage string      Storage backend: local|s3 (default: local)
  -o, --output string       Local output directory (default: ~/.dbvault/backups)
      --bucket string       S3 bucket name (for --storage s3)
      --compress string     Compression: gzip|none (default: gzip)
      --tables strings      Specific tables/collections only (comma-separated)
      --notify              Send Slack notification on completion
```

**Examples:**

```bash
# Full MySQL backup to local disk
dbvault backup --db mysql --host localhost --user root --password secret --name mydb

# PostgreSQL backup to S3 with notification
dbvault backup \
  --db postgres \
  --host db.prod.internal \
  --port 5432 \
  --user postgres \
  --name orders_db \
  --type full \
  --storage s3 \
  --bucket my-backups \
  --notify

# MongoDB incremental backup, specific collections only
dbvault backup \
  --db mongodb \
  --host localhost \
  --name analytics \
  --type incremental \
  --tables "events,sessions"

# SQLite backup (just needs the file path)
dbvault backup --db sqlite --name /var/app/data.db
```

---

### `restore` — Restore from Backup

```bash
dbvault restore [flags]

Flags:
  -d, --db string         Database type: mysql|postgres|mongodb|sqlite (required)
      --host string       Target database host
  -u, --user string       Database username
  -P, --password string   Database password
  -n, --name string       Target database name
  -f, --file string       Backup file path or S3 key (required)
      --storage string    Storage backend where file lives: local|s3
      --tables strings    Restore only specific tables/collections
      --no-verify         Skip checksum verification (not recommended)
      --dry-run           Validate backup file without actually restoring
```

**Examples:**

```bash
# Restore MySQL from local backup
dbvault restore \
  --db mysql \
  --host localhost \
  --user root \
  --password secret \
  --name mydb \
  --file ~/.dbvault/backups/mydb_20260505_020000_full.sql.gz

# Restore specific tables only
dbvault restore \
  --db postgres \
  --name orders_db \
  --file backups/orders_20260505.dump.gz \
  --tables "orders,order_items"

# Restore from S3
dbvault restore \
  --db mysql \
  --name mydb \
  --storage s3 \
  --file backups/mydb_20260505_full.sql.gz
```

---

### `schedule` — Manage Automated Schedules

```bash
dbvault schedule [add|list|remove|pause|resume] [flags]
```

**Sub-commands:**

```bash
# Add a new schedule (cron syntax)
dbvault schedule add \
  --cron "0 2 * * *" \
  --db mysql \
  --host localhost \
  --user root \
  --name mydb \
  --type full \
  --storage s3 \
  --notify \
  --label "nightly-mysql-prod"

# List all active schedules
dbvault schedule list

# Remove a schedule by label or ID
dbvault schedule remove --label "nightly-mysql-prod"

# Start the scheduler daemon (keeps running)
dbvault schedule start

# Run as a one-time cron check (for system cron integration)
dbvault schedule run --label "nightly-mysql-prod"
```

**Schedule List Output:**
```
ID        LABEL                  CRON          DB       STORAGE   LAST RUN              STATUS
────────  ─────────────────────  ────────────  ───────  ────────  ────────────────────  ───────
a3f9c1    nightly-mysql-prod     0 2 * * *     mysql    s3        2026-05-05 02:00:01   ✅ ok
b7e2d4    weekly-postgres-full   0 0 * * 0     postgres local     2026-05-03 00:00:02   ✅ ok
c1a8f6    hourly-mongo-incr      0 * * * *     mongodb  s3        2026-05-05 14:00:01   ❌ fail
```

---

### `test-connection` — Validate Connectivity

```bash
dbvault test-connection [flags]

Flags:
  -d, --db string        Database type (required)
      --host string      Host
  -p, --port int         Port
  -u, --user string      Username
  -P, --password string  Password
  -n, --name string      Database name
```

**Example Output:**
```
Testing connection to MySQL at localhost:3306...

  ✅  TCP reachable          localhost:3306
  ✅  Authentication         root@mydb
  ✅  Database exists        mydb
  ✅  mysqldump available    /usr/bin/mysqldump (v8.0.33)

Connection successful. Ready for backup operations.
```

---

### `config` — Manage Configuration

```bash
# View entire config (with secrets masked)
dbvault config view

# Get a specific key
dbvault config get storage.s3.bucket

# Set a value (persists to config file)
dbvault config set storage.s3.bucket my-new-bucket
dbvault config set notifications.slack.enabled true

# Initialize config with interactive prompts
dbvault config init
```

---

## 📂 Backup Types

### Full Backup
A complete snapshot of the entire database. Supported by all DBMS.

```
Pros:  Simple, self-contained, fastest restore
Cons:  Largest file size, highest DB load during backup
Use:   Nightly or weekly baseline
```

### Incremental Backup (MySQL, MongoDB)
Captures only changes since the **last backup** (full or incremental).

```
MySQL Prerequisites:
  - binary_log = ON in my.cnf
  - log_bin_index path accessible

MongoDB Prerequisites:
  - Replica set with oplog enabled
  - oplogReplay enabled

Pros:  Very small backup files, fast operation
Cons:  Restore requires full + all incremental in sequence
Use:   Hourly between full backups
```

### Differential Backup (PostgreSQL)
Captures all changes since the **last full backup** only.

```
PostgreSQL Prerequisites:
  - archive_mode = on
  - WAL archiving configured

Pros:  Faster restore than incremental (only need full + 1 diff)
Cons:  Grows larger over time between fulls
Use:   Daily between weekly fulls
```

---

## ☁️ Storage Backends

### Local Storage

Files are stored on the local filesystem in the output directory. The directory structure is:

```
~/.dbvault/backups/
└── mysql/
    └── mydb/
        ├── mydb_20260505_020000_full.sql.gz
        ├── mydb_20260505_020000_full.meta.json
        ├── mydb_20260506_020000_full.sql.gz
        └── mydb_20260506_020000_full.meta.json
```

### AWS S3

Requires `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` (or IAM role). Uses the AWS SDK v2 with multipart upload for large files.

```
s3://my-bucket/
└── backups/
    └── mysql/
        └── mydb/
            ├── mydb_20260505_020000_full.sql.gz
            └── mydb_20260505_020000_full.meta.json
```

### Backup Metadata File (`.meta.json`)

Every backup generates a sidecar metadata file:

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "db_type": "mysql",
  "db_host": "localhost",
  "db_name": "mydb",
  "backup_type": "full",
  "compression": "gzip",
  "storage_type": "s3",
  "storage_path": "backups/mysql/mydb/mydb_20260505_020000_full.sql.gz",
  "start_time": "2026-05-05T02:00:00Z",
  "end_time": "2026-05-05T02:04:37Z",
  "duration_seconds": 277,
  "original_size_bytes": 524288000,
  "compressed_size_bytes": 98566144,
  "compression_ratio": "81.2%",
  "checksum_sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "status": "success",
  "dbvault_version": "1.0.0"
}
```

---

## 📅 Scheduling

DBVault uses standard **cron syntax** (6-field with seconds support):

```
┌─────────── second (0-59)    [optional]
│ ┌───────── minute (0-59)
│ │ ┌─────── hour (0-23)
│ │ │ ┌───── day of month (1-31)
│ │ │ │ ┌─── month (1-12)
│ │ │ │ │ ┌─ day of week (0-6, Sun=0)
│ │ │ │ │ │
* * * * * *
```

**Common Patterns:**

| Cron Expression | Meaning |
|---|---|
| `0 2 * * *` | Daily at 02:00 AM |
| `0 0 * * 0` | Weekly, Sunday midnight |
| `0 * * * *` | Every hour |
| `0 2 1 * *` | Monthly, 1st at 02:00 AM |
| `*/30 * * * *` | Every 30 minutes |

Schedules are persisted in `~/.dbvault/schedules.json` and survive restarts.

---

## 📋 Logging

All backup operations are logged in structured JSON format:

```json
{
  "time": "2026-05-05T02:04:37.123Z",
  "level": "INFO",
  "msg": "backup completed",
  "backup_id": "a1b2c3d4",
  "db_type": "mysql",
  "db_name": "mydb",
  "backup_type": "full",
  "duration_ms": 277423,
  "compressed_size_bytes": 98566144,
  "storage": "s3",
  "path": "backups/mysql/mydb/mydb_20260505_full.sql.gz",
  "status": "success"
}
```

Log files rotate automatically at 100MB, keeping the last 7 rotations.

**Log location:** `~/.dbvault/logs/dbvault.log`

---

## 🔔 Notifications

### Slack

Set your webhook URL in config or env var:

```bash
dbvault config set notifications.slack.webhook_url https://hooks.slack.com/services/...
dbvault config set notifications.slack.enabled true
```

**Slack Message Format:**

```
🛡️ DBVault Backup Report

Status:      ✅ Success
Database:    mysql / mydb (localhost)
Type:        Full Backup
Duration:    4m 37s
Size:        500 MB → 94 MB (gzip, 81.2% reduction)
Stored at:   s3://my-bucket/backups/mysql/mydb/mydb_20260505_full.sql.gz
Completed:   2026-05-05 02:04:37 UTC
```

For failures, the message includes the error message and last log lines.

---

## ♻️ Restore Operations

### Full Restore

```bash
dbvault restore --db mysql --name mydb --file backup.sql.gz
```

Flow: Download → Verify checksum → Decompress → Pipe to `mysql` CLI → Log

### Selective Restore

Restore specific tables without touching the rest of the database:

```bash
# MySQL / PostgreSQL: restore only the users and sessions tables
dbvault restore \
  --db postgres \
  --name mydb \
  --file backup.dump.gz \
  --tables "users,sessions"

# MongoDB: restore only specific collections
dbvault restore \
  --db mongodb \
  --name mydb \
  --file backup.archive.gz \
  --tables "users,events"
```

---

## 🧠 Design Decisions

| Decision | Rationale |
|---|---|
| **Shell out to native DB tools** | mysqldump/pg_dump handle edge cases, large tables, locks, and formats better than any pure-Go reimplementation |
| **Streaming compression** | Compressor wraps the subprocess stdout as `io.Reader` — avoids holding the full uncompressed dump in memory |
| **Pluggable storage via interface** | `Storage` interface makes adding GCS or Azure a matter of implementing 5 methods, with zero changes to backup logic |
| **Sidecar `.meta.json` files** | Keeps backup archives self-contained and queryable without a separate database. Checksum validates integrity before restore |
| **Cron over OS scheduler** | Ships as a self-contained daemon with `robfig/cron` for portability across Linux, macOS, and Windows without OS-specific setup |
| **Viper for config** | Handles the flags > env > file > default priority chain automatically. Avoids duplicating validation logic |
| **`slog` over third-party logger** | Standard library (Go 1.21+) — no extra dependency, structured JSON out of the box, sufficient for this tool |

---

## 🔒 Security Considerations

- **Passwords** are never written to log files or metadata files
- **Env vars** are preferred over CLI flags for secrets (avoids shell history exposure)
- **Checksums** (SHA-256) are verified before every restore operation
- **S3 uploads** use server-side encryption (SSE-S3) by default
- Backup files on local disk are created with **mode 0600** (owner read/write only)

---

## 🗺️ Roadmap

- [ ] Google Cloud Storage backend
- [ ] Azure Blob Storage backend
- [ ] Retention policy (auto-delete backups older than N days)
- [ ] Email notifications (SMTP)
- [ ] `dbvault list` — list all backups with size, date, type
- [ ] `dbvault verify` — verify all backups in a directory
- [ ] TUI dashboard (Bubble Tea)
- [ ] Docker image for containerized deployments
- [ ] Prometheus metrics endpoint for backup monitoring

---

## 🛠️ Makefile Targets

```bash
make build        # Build binary to ./bin/dbvault
make install      # Install to $GOPATH/bin
make test         # Run all tests
make lint         # Run golangci-lint
make clean        # Remove build artifacts
make release      # Cross-compile for linux/darwin/windows (amd64 + arm64)
```

---

## 📄 License

MIT — see [LICENSE](LICENSE)

---

<p align="center">Built with ❤️ in Go · by Ritik Kumar</p>