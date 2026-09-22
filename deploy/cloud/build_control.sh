#!/bin/sh
# Build a prepared context on slang-app and activate its Compose services.
set -eu
context=${1:?Usage: build_control.sh /absolute/context release-tag}
tag=${2:?Specify a release tag, or working for development}
case "$context" in /srv/slang/build/*) ;; *) echo 'Context must be below /srv/slang/build' >&2; exit 1;; esac
case "$tag" in *[!a-zA-Z0-9_.-]*|'') echo 'Invalid image tag' >&2; exit 1;; esac
if [ "$tag" != working ]; then
  python3 -c 'import json,sys; m=json.load(open(sys.argv[1])); assert not any(v["dirty"] for v in m.values()), "Release sources must be clean"' "$context/source-manifest.json"
  if docker image inspect "slang-auth:$tag" >/dev/null 2>&1; then
    echo 'Release tag already exists; choose a new tag' >&2; exit 1
  fi
fi
for service in auth repo deployment usage vision search meta transit customer narrow; do
  revision=$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))[sys.argv[2]]["revision"])' "$context/source-manifest.json" "$service")
  docker build --progress=plain --label "org.opencontainers.image.revision=$revision" \
    -t "slang-$service:$tag" --build-arg "SERVICE=$service" \
    -f "$context/cloud/Dockerfile.service" "$context" >"$context/$service-build.log" 2>&1 || {
      tail -40 "$context/$service-build.log"; exit 1;
    }
  echo "Built $service $revision"
done
docker build -t "slang-vision-renderer:$tag" -f "$context/cloud/Dockerfile.vision-renderer" "$context" >"$context/renderer-build.log" 2>&1 || {
  tail -40 "$context/renderer-build.log"; exit 1;
}
if [ -f /srv/slang/compose.control.yml ]; then
  cp /srv/slang/compose.control.yml "$context/compose.control.previous.yml"
fi
sed "s/:working/:$tag/g" "$context/cloud/compose.control.yml" >/srv/slang/compose.control.yml
install -m 644 "$context/cloud/compose.data.yml" /srv/slang/compose.data.yml
compose() { docker compose -f /srv/slang/compose.data.yml -f /srv/slang/compose.control.yml "$@"; }
compose config --quiet
for service in auth repo deployment usage vision customer narrow; do
  compose run --rm --no-deps -e INIT_SCHEMA=true "$service" python -c 'import index; print("Schema ready")'
done
compose up -d
install -m 644 "$context/cloud/slang-maintenance.service" "$context/cloud/slang-maintenance.timer" /etc/systemd/system/
systemctl daemon-reload
systemctl enable --now slang-maintenance.timer
python3 "$context/cloud/record_release.py" "$context/source-manifest.json" "$tag"
