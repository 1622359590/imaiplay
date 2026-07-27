#!/usr/bin/env bash
set -Eeuo pipefail

if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
  echo "Usage: $0 BACKUP_FILE [DATABASE]" >&2
  exit 2
fi

backup_file="$1"
database="${2:-${PGDATABASE:-imaiplay}}"
if [ ! -f "$backup_file" ]; then
  echo "Backup file not found: $backup_file" >&2
  exit 2
fi
if ! command -v pg_restore >/dev/null 2>&1 && [ "${DRY_RUN:-0}" != "1" ]; then
  echo "pg_restore is required but was not found" >&2
  exit 127
fi

if [ "${DRY_RUN:-0}" = "1" ]; then
  printf 'pg_restore --clean --if-exists --no-owner --dbname=%q %q\n' "$database" "$backup_file"
  exit 0
fi

echo "This will replace objects in database '$database' using '$backup_file'."
read -r -p "Type RESTORE to continue: " confirmation
if [ "$confirmation" != "RESTORE" ]; then
  echo "Restore cancelled"
  exit 1
fi

pg_restore --clean --if-exists --no-owner --dbname="$database" "$backup_file"
echo "Restore complete"
