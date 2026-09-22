#!/bin/sh
# Run as root with this repository checked out at /srv/tryslang/src.
# Cloudflare DNS/redirects and the provider firewall are managed separately.
set -eu
ROOT=/srv/tryslang
SOURCE="$ROOT/src/deploy/public"
apt-get update -qq
DEBIAN_FRONTEND=noninteractive apt-get install -y docker.io docker-compose-v2 caddy python3
if ! id tryslang >/dev/null 2>&1; then
  useradd --system --uid 65532 --user-group --home-dir "$ROOT" --shell /usr/sbin/nologin tryslang
fi
usermod -aG docker tryslang
systemctl enable --now docker
mkdir -p "$ROOT/backups" "$ROOT/state"
chown tryslang:tryslang "$ROOT/state"
chmod 700 "$ROOT/state"
python3 "$SOURCE/prepare.py" "$ROOT"
cp "$SOURCE/gateway.py" "$ROOT/gateway.py"
docker build -t tryslang-runtime:current -f "$SOURCE/Dockerfile" "$ROOT/src"
if ! docker network inspect tryslang-sessions >/dev/null 2>&1; then
  docker network create --internal tryslang-sessions
fi
if [ ! -f "$ROOT/backups/Caddyfile.original" ]; then
  cp /etc/caddy/Caddyfile "$ROOT/backups/Caddyfile.original"
fi
cp "$SOURCE/tryslang.service" /etc/systemd/system/tryslang.service
cp "$SOURCE/Caddyfile" /etc/caddy/Caddyfile
caddy validate --config /etc/caddy/Caddyfile
systemctl daemon-reload
systemctl enable --now tryslang caddy
systemctl restart tryslang
systemctl reload caddy
python3 "$SOURCE/smoke.py" http://127.0.0.1:5150
