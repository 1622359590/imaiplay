#!/bin/sh
set -eu

release_image=${1:?release image is required}
deploy_log=$(mktemp)
trap 'rm -f "$deploy_log"' EXIT

deploy_once() {
  IMAI_PLAY_IMAGE="$release_image" docker compose \
    -f docker-compose.yml \
    -f docker-compose.bt.yml \
    -f docker-compose.deploy.yml \
    up -d --no-build --remove-orphans
}

if deploy_once >"$deploy_log" 2>&1; then
  cat "$deploy_log"
  exit 0
fi

cat "$deploy_log" >&2
if ! grep -Eq 'iptables.*No chain/target/match by that name' "$deploy_log"; then
  exit 1
fi

echo 'Docker iptables chain is missing; restarting Docker and retrying once.' >&2
if command -v systemctl >/dev/null 2>&1; then
  systemctl restart docker
elif command -v service >/dev/null 2>&1; then
  service docker restart
else
  echo 'Cannot restart Docker: neither systemctl nor service is available.' >&2
  exit 1
fi

attempt=1
while ! docker info >/dev/null 2>&1; do
  if [ "$attempt" -ge 15 ]; then
    echo 'Docker did not become ready after restart.' >&2
    exit 1
  fi
  attempt=$((attempt + 1))
  sleep 2
done

deploy_once
