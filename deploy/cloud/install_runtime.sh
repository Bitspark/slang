#!/bin/sh
# Run as root on slang-runtime after copying the Linux binary and cloud sources.
set -eu
binary=${1:?Usage: install_runtime.sh /path/slang /path/cloud release-tag source-revision}
cloud=${2:?Cloud source directory required}
tag=${3:?Image tag required}
revision=${4:?Source revision required}
case "$tag" in *[!a-zA-Z0-9_.-]*|'') exit 1;; esac
install -d -m 755 /srv/slang/runtime /srv/slang/runtime/tls /srv/slang/runner /srv/slang/ops
install -m 755 "$binary" /srv/slang/runner/slang
install -m 644 "$cloud/Dockerfile.runner" /srv/slang/runner/Dockerfile
docker build --label "org.opencontainers.image.revision=$revision" -t "slang-runner:$tag" /srv/slang/runner
id slang-router >/dev/null 2>&1 || useradd --system --no-create-home --shell /usr/sbin/nologin slang-router
python3 -m venv /srv/slang/runtime/venv
/srv/slang/runtime/venv/bin/pip install -r "$cloud/runtime_requirements.txt"
install -m 644 "$cloud"/runtime_*.py "$cloud"/runtime_*.sh /srv/slang/runtime/
python3 "$cloud/initialize_runtime.py"
if [ ! -f /srv/slang/runtime/tls/server.key ]; then
  umask 077
  openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:3072 -out /srv/slang/runtime/tls/server.key 2>/dev/null
fi
openssl req -new -key /srv/slang/runtime/tls/server.key -out /tmp/slang-runtime.csr -subj '/CN=slang-runtime'
chmod 644 /tmp/slang-runtime.csr
install -d /etc/systemd/system/docker.service.d
install -m 644 "$cloud/docker-slang-firewall.conf" /etc/systemd/system/docker.service.d/slang-firewall.conf
install -m 644 "$cloud"/slang-runtime-*.service /etc/systemd/system/
printf 'SLANG_RUNNER_IMAGE=slang-runner:%s\n' "$tag" >/srv/slang/runtime.env
systemctl daemon-reload
systemctl enable --now slang-runtime-network
if [ -f /srv/slang/runtime/tls/server.crt ]; then
  systemctl enable slang-runtime-agent slang-runtime-router
  systemctl restart slang-runtime-agent slang-runtime-router
else
  echo 'Issue the CSR on slang-app and install the public certificates before starting runtime services.'
fi
python3 - "$revision" "$tag" <<'PY'
import hashlib,json,pathlib,subprocess,sys
root=pathlib.Path('/srv/slang')
image='slang-runner:'+sys.argv[2]
manifest={'revision':sys.argv[1], 'image':image, 'image_id':subprocess.check_output(['docker','inspect','--format','{{.Id}}',image],text=True).strip(),
          'binary_sha256':hashlib.sha256((root/'runner/slang').read_bytes()).hexdigest()}
(root/'runtime-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
PY
