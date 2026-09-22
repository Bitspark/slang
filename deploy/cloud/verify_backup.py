"""Controller-side recovery check; decrypt only in memory, restore to isolated Docker.

Requires age and the Slang SSH aliases. The private identity never leaves the
controller. Restored databases have no network or published ports and are removed
when the test ends. Does not alter the production databases.
"""
import argparse
import io
import json
import sqlite3
import subprocess
import tarfile
import time


def ssh(host, command, data=None, check=True):
    return subprocess.run(['ssh', host, command], input=data, capture_output=True, check=check)


def archive(host, identity, age):
    name = ssh(host, 'sudo find /srv/slang/backups/local -maxdepth 1 -name "*.tar.gz.age" -printf "%f\\n" | sort | tail -1').stdout.decode().strip()
    assert name.startswith(host + '-') and '/' not in name
    encrypted = ssh(host, 'sudo cat /srv/slang/backups/local/' + name).stdout
    plaintext = subprocess.run([age, '-d', '-i', identity], input=encrypted, capture_output=True, check=True).stdout
    return name, tarfile.open(fileobj=io.BytesIO(plaintext), mode='r:gz')


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--identity', required=True)
    parser.add_argument('--age', default='age')
    args = parser.parse_args()
    app_name, app = archive('slang-app', args.identity, args.age)
    runtime_name, runtime = archive('slang-runtime', args.identity, args.age)
    host = 'slang-runtime'
    pg = 'slang-restore-check-postgres'
    redis = 'slang-restore-check-redis'
    # Refuse to replace any pre-existing container, including an interrupted test.
    for name in (pg, redis):
        assert ssh(host, 'sudo docker inspect ' + name, check=False).returncode != 0, 'Restore-check container already exists'
    try:
        ssh(host, 'sudo docker run -d --name ' + pg + ' --network none --memory 512m --pids-limit 64 --tmpfs /var/lib/postgresql/data:rw,size=256m -e POSTGRES_HOST_AUTH_METHOD=trust postgres:17-bookworm')
        for _ in range(30):
            if ssh(host, 'sudo docker exec ' + pg + ' pg_isready -U postgres', check=False).returncode == 0:
                break
            time.sleep(1)
        dump = app.extractfile('backup/postgres.sql').read()
        ssh(host, 'sudo docker exec -i ' + pg + ' psql -U postgres -d postgres -v ON_ERROR_STOP=1', dump)
        databases = ssh(host, 'sudo docker exec ' + pg + ' psql -U postgres -At -c "SELECT datname FROM pg_database"').stdout.decode().splitlines()
        expected = {'slang_' + s for s in ('auth', 'repo', 'deployment', 'usage', 'vision', 'customer', 'narrow')}
        assert expected.issubset(databases)
        counts = {}
        for database, table in [('slang_auth', 'user'), ('slang_repo', 'bundle'), ('slang_deployment', 'instance')]:
            sql = ('SELECT count(*) FROM "' + table + '";').encode()
            counts[database] = int(ssh(host, 'sudo docker exec -i ' + pg + ' psql -U postgres -d ' + database + ' -At', sql).stdout)
        assert counts['slang_auth'] > 0 and counts['slang_repo'] >= 122
        ssh(host, 'sudo docker run -d --name ' + redis + ' --network none --memory 128m --pids-limit 32 --tmpfs /data:rw,size=64m redis:7.2-bookworm sleep 300')
        # Feed RDB to a fixed path in the isolated test container.
        ssh(host, 'sudo docker exec -i ' + redis + ' sh -c "cat > /data/dump.rdb"', app.extractfile('backup/dump.rdb').read())
        ssh(host, 'sudo docker exec ' + redis + ' redis-server --daemonize yes --appendonly no --dir /data --dbfilename dump.rdb')
        size = int(ssh(host, 'sudo docker exec ' + redis + ' redis-cli DBSIZE').stdout)
        assert size > 0
        state = sqlite3.connect(':memory:')
        state.deserialize(runtime.extractfile('backup/runtime-state/state.db').read())
        assert state.execute('PRAGMA integrity_check').fetchone()[0] == 'ok'
        rows = state.execute('SELECT iid FROM instance').fetchall()
        for (iid,) in rows:
            assert len(iid) == 32
            json.load(runtime.extractfile('backup/runtime-state/instances/' + iid + '.json'))
        events = sqlite3.connect(':memory:')
        events.deserialize(runtime.extractfile('backup/runtime-state/telemetry/events.db').read())
        assert events.execute('PRAGMA integrity_check').fetchone()[0] == 'ok'
        print(json.dumps({'result': 'passed', 'backups': [app_name, runtime_name], 'restored_databases': sorted(expected),
            'row_counts': counts, 'redis_keys': size, 'runtime_instances': len(rows)}))
    finally:
        for name in (pg, redis):
            ssh(host, 'sudo docker rm -f ' + name, check=False)


if __name__ == '__main__':
    main()
