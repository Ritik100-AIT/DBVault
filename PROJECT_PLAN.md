# DBVault Project Plan

## Overview
This project plan tracks the current DBVault CLI implementation, completed work, and next priorities.

## Current Status
- Cobra CLI is wired for `backup`, `restore`, `test-connection`, `schedule add`, `schedule list`, `schedule remove`, and `config view`.
- Viper config loading is implemented with defaults, config file support, environment variable binding, and CLI flag override support.
- Internal architecture exists for:
  - `internal/config`
  - `internal/db` connector interface and stub connectors
  - `internal/storage` interface with local backend
  - `internal/scheduler` persisted schedule store
  - `internal/backup` manager orchestration
  - `internal/compress` gzip pipeline
- Tests added for:
  - config merge behavior
  - CLI command flow
  - storage backend save/load
  - scheduler persistence
- Code comments added for clarity on major functions and modules.

## Completion Estimate
- Current completion: ~60-70% of the core CLI foundation.
- Remaining work is focused on real backend implementation, scheduler execution, and polish.

## Priority Tasks
1. Real database connector implementation
   - `mysqldump` / `mysql`
   - `pg_dump` / `psql`
   - `mongodump` / `mongorestore`
   - SQLite file backup/restore
2. Storage backend improvements
   - implement S3 backend
   - support remote/cloud storage options
   - handle storage errors and retries
3. Scheduler execution
   - hook scheduled jobs into backup manager
   - use cron scheduling or background runner
   - support persistent schedule execution
4. Restore flow tests and CLI coverage
   - add restore command tests
   - add schedule command tests
   - verify config/env/flag merge behavior end to end
5. Documentation and polish
   - update README with real usage examples
   - document env vars, config file format, and commands
   - ensure help text is complete
6. Logging and notifications
   - integrate structured logging
   - add Slack notifications if needed

## Future Milestones
- **M1**: Stable local backup/restore CLI with config and env support
- **M2**: Scheduler persistence + execution working
- **M3**: Cloud storage support (S3)
- **M4**: Production-ready logging, notifications, and error handling

## Notes
- Use `go test ./...` after each change.
- Keep CLI command behavior deterministic and testable.
- Prefer small, isolated changes so that each feature can be reviewed and validated.
