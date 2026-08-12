#!/bin/sh
set -eu

health_url="${1:-http://127.0.0.1:18080/health/db}"
retries="${DEPLOY_HEALTH_RETRIES:-10}"
retry_delay="${DEPLOY_HEALTH_RETRY_DELAY_SECONDS:-2}"
max_time="${DEPLOY_HEALTH_MAX_TIME_SECONDS:-10}"

attempt=0
while [ "$attempt" -le "$retries" ]; do
  if curl \
    --fail \
    --silent \
    --show-error \
    --max-time "$max_time" \
    "$health_url"
  then
    exit 0
  fi

  attempt=$((attempt + 1))
  if [ "$attempt" -le "$retries" ] && [ "$retry_delay" -gt 0 ]; then
    sleep "$retry_delay"
  fi
done

exit 1
