#!/bin/sh
set -eu

health_url="${1:-http://127.0.0.1:18080/health/db}"
retries="${DEPLOY_HEALTH_RETRIES:-10}"
retry_delay="${DEPLOY_HEALTH_RETRY_DELAY_SECONDS:-2}"
max_time="${DEPLOY_HEALTH_MAX_TIME_SECONDS:-10}"

curl \
  --fail \
  --silent \
  --show-error \
  --retry "$retries" \
  --retry-all-errors \
  --retry-delay "$retry_delay" \
  --max-time "$max_time" \
  "$health_url"
