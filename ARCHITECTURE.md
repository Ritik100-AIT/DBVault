# DBVault CLI Architecture and Code Flow

## Overview
DBVault is a Go-based CLI tool for database backup, restore, and scheduling operations. It uses Cobra for command parsing, Viper for configuration management, and a clean architecture with interfaces for extensibility.

## High-Level Flow
1. **Entry Point**: `main.go` calls `cmd.Execute()` to start the CLI.
2. **Command Parsing**: Cobra handles commands (e.g., `backup`, `restore`) with flags and subcommands.
3. **Configuration Loading**: Viper merges defaults, config file, env vars, and CLI flags via `internal/config`.
4. **Execution**: Commands instantiate managers (e.g., `BackupManager`) that orchestrate DB connectors, storage backends, and compression.
5. **Persistence**: Scheduler uses JSON files for schedule storage; storage backends handle file I/O.

## Key Components
- **cmd/**: CLI layer with Cobra commands (root, backup, restore, etc.). Each command has `RunE` for execution logic.
- **internal/config/**: Viper setup and config loading/merging.
- **internal/db/**: DBConnector interface and stubs for MySQL, PostgreSQL, MongoDB, SQLite.
- **internal/backup/**: BackupManager orchestrates backup flow (test connection → backup → compress → save).
- **internal/storage/**: StorageBackend interface with local and S3 implementations.
- **internal/scheduler/**: Scheduler for adding/listing/removing schedules with JSON persistence.
- **internal/compress/**: Compression stubs (GzipCompress/Decompress).
- **internal/models/**: Shared structs (AppConfig, DBConfig, etc.).
- **internal/logger/**: Logging utilities (placeholder).
- **internal/notify/**: Notification stubs (e.g., Slack).

## Code-Level Understanding
- **Interfaces for Extensibility**: DBConnector, StorageBackend allow easy swapping of implementations.
- **Error Handling**: All functions return errors; commands use `RunE` for proper error propagation.
- **Testing**: Unit tests in each package validate logic (e.g., config merging, manager execution).
- **Config Precedence**: CLI flags > env vars > config file > defaults (handled by Viper).
- **Flow Example (Backup)**: Command loads config → creates connector/backend → manager runs (test conn, backup, compress, save).

## Current Status
- CLI wiring and config merging: Complete.
- Internal stubs and placeholders: Complete.
- Real DB logic, S3 storage, scheduler execution: Pending.
- Tests and comments: Added for reference.

For details, see PROJECT_PLAN.md or code comments.