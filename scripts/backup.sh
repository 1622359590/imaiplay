#!/usr/bin/env bash
set -Eeuo pipefail

BACKUP_DIR="${BACKUP_DIR:-./backups}"
RETENTION_COUNT="${RETENTION_COUNT:-7}"
DATABASE="${PGDATABASE:-imaiplay}"

if ! [[ "$RETENTION_COUNT" =~ ^[0-9]+$ ]] || [ "$RETENTION_COUNT" -lt 1 ]; then
  echo "RETENTION_COUNT must be a positive integer" >&2
  exit 2
fi
mkdir -p "$BACKUP_DIR"
timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
output="$BACKUP_DIR/imaiplay_${timestamp}.dump"

if [ "${DRY_RUN:-0}" = "1" ]; then
  printf 'pg_dump --format=custom --file=%q %q\n' "$output" "$DATABASE"
  printf 'retain newest %s dump files in %q\n' "$RETENTION_COUNT" "$BACKUP_DIR"
  exit 0
fi

if ! command -v pg_dump >/dev/null 2>&1; then
  echo "pg_dump is required but was not found" >&2
  exit 127
fi

echo "Backing up database $DATABASE to $output"
pg_dump --format=custom --no-owner --file="$output" "$DATABASE"

mapfile_cmd=""
while IFS= read -r old_backup; do
  [ -n "$old_backup" ] || continue
  rm -- "$old_backup"
  echo "Removed old backup: $old_backup"
done < <(find "$BACKUP_DIR" -maxdepth 1 -type f -name 'imaiplay_*.dump' -print | sort -r | tail -n +$((RETENTION_COUNT + 1)))

echo "Backup complete: $output"
