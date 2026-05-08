# MinIO Backup & Restore Demo Guide

This guide shows how to demo DBVault backup and restore using MinIO and SQLite.
It is the fastest way to show MinIO storage without requiring a full database server.

## Prerequisites

- `go` installed
- `sqlite3` installed
- `docker` installed (for MinIO)
- `mc` (MinIO client) is optional but helpful
- your DBVault project directory is `/Users/ritik_kumar/Desktop/DBVault`

## 1. Start MinIO locally

```bash
cd /Users/ritik_kumar/Desktop/DBVault

docker run -d --name minio-demo -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  quay.io/minio/minio server /data --console-address ":9001"
```

Wait a few seconds, then open http://127.0.0.1:9001 in your browser.
Login with:

- Access key: `minioadmin`
- Secret key: `minioadmin`

## 2. Create a bucket for backups

Use the MinIO console or `mc`:

```bash
mc alias set local http://127.0.0.1:9000 minioadmin minioadmin
mc mb local/dbvault-backups
```

If you do not have `mc`, the web UI is enough.

## 3. Prepare a demo SQLite database

```bash
cd /Users/ritik_kumar/Desktop/DBVault
rm -f demo.db restored.db
sqlite3 demo.db <<'SQL'
CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT);
INSERT INTO users (name) VALUES ('Alice'), ('Bob'), ('Carol');
SQL

sqlite3 demo.db 'SELECT * FROM users;'
```

You should see the demo rows printed.

## 4. Create a DBVault config file for MinIO

Save this as `config.yaml` in the project root:

```yaml
database:
  type: sqlite
  name: demo.db

backup:
  type: full
  compression: gzip

storage:
  type: s3
  s3:
    bucket: dbvault-backups
    region: us-east-1
    endpoint: http://127.0.0.1:9000
    force_path_style: true
    access_key: minioadmin
    secret_key: minioadmin
    prefix: demo

notifications:
  slack:
    enabled: false
```

## 5. Run a backup to MinIO

```bash
go run . --config config.yaml backup --db sqlite --name demo.db --storage s3 --compress gzip
```

Expected output should include a line like:

```
Backup stored at s3://dbvault-backups/demo/sqlite-demo.db-20260508T...Z.gz
```

## 6. Verify the backup in MinIO

Open the MinIO console or use `mc`:

```bash
mc ls local/dbvault-backups/demo
```

You should see both:

- `sqlite-demo.db-...Z.gz`
- `sqlite-demo.db-...Z.gz.meta.json`

The `.meta.json` file is the backup metadata sidecar.

## 7. Restore the backup from MinIO

Choose the backup path printed by the backup command. Example:

```bash
go run . --config config.yaml restore --source sqlite-demo.db-20260508T150405Z.gz --db sqlite --name restored.db --storage s3 --compress gzip
```

If you use multiple backups, supply the exact `s3://...` path or object key.

## 8. Confirm the restore result

```bash
sqlite3 restored.db 'SELECT * FROM users;'
```

You should see the same rows:

- Alice
- Bob
- Carol

## 9. Clean up when finished

```bash
docker rm -f minio-demo
rm -f demo.db restored.db
```

## Notes for your demo

- Use `config.yaml` so the same settings work for both backup and restore.
- Emphasize:
  - MinIO is compatible with S3 mode
  - DBVault writes both archive and `.meta.json` checksum metadata
  - Restore verifies checksum before replaying the backup
- If you want, you can also demo a second restore by removing `restored.db` and running restore again.
