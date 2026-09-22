"""Encrypted host backup with a restricted SSH transfer to the other Slang VM.

The age private identity lives only on the controller. Configuration and the
public recipient live in /srv/slang/backup.json; no secrets are logged.
"""
import contextlib
import datetime
import fcntl
import json
import os
from pathlib import Path
import shutil
import sqlite3
import subprocess
import tarfile
import tempfile

ROOT = Path('/srv/slang')


def run(*args, **kwargs):
    return subprocess.run(args, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, **kwargs)


def main():
    os.umask(0o077)
    config = json.loads((ROOT / 'backup.json').read_text())
    host = config['host']
    if host not in ('slang-app', 'slang-runtime'):
        raise ValueError('Unknown host')
    out = ROOT / 'backups/local'
    out.mkdir(parents=True, exist_ok=True)
    with (ROOT / 'backup.lock').open('w') as lock:
        fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        with tempfile.TemporaryDirectory(prefix='slang-backup-', dir='/var/tmp') as temp:
            stage = Path(temp)
            compose = ['docker', 'compose', '-f', str(ROOT / 'compose.data.yml'), '-f', str(ROOT / 'compose.control.yml')]
            stopped = []
            try:
                if host == 'slang-app':
                    services = run(*compose, 'ps', '--services', '--status', 'running').stdout.decode().split()
                    stopped = [s for s in services if s not in ('postgres', 'redis')]
                    if stopped:
                        run(*compose, 'stop', '--timeout', '20', *stopped)
                    with (stage / 'postgres.sql').open('wb') as output:
                        subprocess.run(['docker', 'exec', 'slang-postgres-1', 'pg_dumpall', '-U', 'slang'], stdout=output, stderr=subprocess.PIPE, check=True)
                    run('docker', 'exec', 'slang-redis-1', 'redis-cli', 'SAVE')
                    run('docker', 'cp', 'slang-redis-1:/data/dump.rdb', str(stage / 'dump.rdb'))
                else:
                    stopped = ['slang-runtime-agent', 'slang-runtime-router']
                    run('systemctl', 'stop', *stopped)
                    state = Path('/var/lib/slang-runtime')
                    shutil.copytree(state, stage / 'runtime-state', ignore=shutil.ignore_patterns('*-wal', '*-shm', 'locks'))
                    for relative in ('state.db', 'telemetry/events.db'):
                        with contextlib.closing(sqlite3.connect(state / relative)) as source, contextlib.closing(sqlite3.connect(stage / 'runtime-state' / relative)) as target:
                            source.backup(target)
                # Immutable sources/artifacts are rebuilt from the release manifest.
                for name in ('secrets', 'runtime/tls'):
                    source = ROOT / name
                    if source.exists():
                        shutil.copytree(source, stage / 'srv/slang' / name)
                for pattern in ('*.env', 'compose.*.yml', '*manifest.json', 'backup.json'):
                    for source in ROOT.glob(pattern):
                        dest = stage / 'srv/slang' / source.name
                        dest.parent.mkdir(parents=True, exist_ok=True)
                        shutil.copy2(source, dest)
                shutil.copy2('/etc/caddy/Caddyfile', stage / 'Caddyfile')
                (stage / 'backup-info.json').write_text(json.dumps({'host': host, 'created_at': datetime.datetime.now(datetime.timezone.utc).isoformat(), 'format': 1}))
            finally:
                if stopped:
                    if host == 'slang-app':
                        run(*compose, 'start', *stopped)
                    else:
                        run('systemctl', 'start', *stopped)
            timestamp = datetime.datetime.now(datetime.timezone.utc).strftime('%Y%m%dT%H%M%SZ')
            target = out / f'{host}-{timestamp}.tar.gz.age'
            # Stream the archive into encryption; never persist a plaintext tar.
            with target.with_suffix('.partial').open('wb') as output:
                process = subprocess.Popen(['age', '-r', config['recipient']], stdin=subprocess.PIPE, stdout=output, stderr=subprocess.PIPE)
                with tarfile.open(fileobj=process.stdin, mode='w|gz') as archive:
                    archive.add(stage, arcname='backup')
                process.stdin.close()
                error = process.stderr.read()
                if process.wait():
                    raise RuntimeError('Backup encryption failed')
            target.with_suffix('.partial').replace(target)
            run('rsync', '-t', '--chmod=F600', '-e',
                'ssh -i /srv/slang/secrets/backup-transfer -o IdentitiesOnly=yes -o StrictHostKeyChecking=yes -o UserKnownHostsFile=/srv/slang/secrets/backup-known-hosts',
                str(target), f"slang-backup@{config['peer']}:{target.name}")
            # Keep seven successful daily snapshots on each host. Peer pruning
            # runs as root here, outside the restricted receiving SSH account.
            for directory in (out, ROOT / 'backups/peer'):
                for prefix in ('slang-app', 'slang-runtime'):
                    for old in sorted(directory.glob(prefix + '-*.tar.gz.age'))[:-7]:
                        old.unlink()
            print(json.dumps({'backup': target.name, 'encrypted_bytes': target.stat().st_size, 'peer_copy': 'verified by rsync'}))


if __name__ == '__main__':
    main()
