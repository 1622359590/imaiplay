#!/bin/sh
set -eu

PROJECT_DIR="${IMAIPLAY_PROJECT_DIR:-/www/wwwroot/imaiplay}"
LOCK_FILE="${IMAIPLAY_DEPLOY_LOCK:-/tmp/imaiplay-deploy.lock}"

exec 9>"$LOCK_FILE"
flock -n 9 || exit 0

cd "$PROJECT_DIR"
git fetch origin main

if git diff --quiet HEAD origin/main; then
    exit 0
fi

git merge --ff-only origin/main
docker compose -f docker-compose.bt.yml up -d --build --remove-orphans
curl --fail --silent --show-error --retry 10 --retry-delay 2 http://127.0.0.1:18080/health
