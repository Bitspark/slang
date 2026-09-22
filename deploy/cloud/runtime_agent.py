"""Private mTLS management API. This is the only Slang service with Docker access.

Clients can supply bundle data, IDs and one of three modes, never Docker flags,
images, commands or host paths. TLS client verification is enforced by Gunicorn.
"""
import contextlib
import datetime
import hashlib
import json
import os
import shutil
import subprocess
import time
import uuid

from flask import Flask, Response, abort, jsonify, request
from runtime_common import ROOT, ID, connection, event, invoke, record

app = Flask(__name__)
app.config['MAX_CONTENT_LENGTH'] = 8 * 1024 * 1024
IMAGE = os.environ.get('SLANG_RUNNER_IMAGE', 'slang-runner:working')


def docker(*args, check=True):
    result = subprocess.run(['docker', *args], capture_output=True, timeout=30)
    if check and result.returncode:
        raise RuntimeError('Docker operation failed: ' + result.stderr.decode(errors='replace')[:500])
    return result


def owner(value):
    if value != 'system' and (not isinstance(value, str) or not ID.fullmatch(value)):
        abort(400, 'Invalid owner')
    return value


def identity(iid):
    if not ID.fullmatch(iid):
        abort(400, 'Invalid deployment ID')
    return 'slang-program-' + iid


def owned(iid):
    identity(iid)
    value = request.get_json(silent=True) or request.args
    item = record(iid)
    if not item or item['owner'] != owner(value.get('owner')):
        abort(404)
    return item


def snapshot(definition):
    try:
        bid = str(uuid.UUID(definition['main']))
        assert bid in definition['blueprints']
    except (KeyError, ValueError, TypeError, AttributeError, AssertionError):
        abort(400, 'Invalid bundle')
    data = json.dumps(definition, separators=(',', ':')).encode()
    if len(data) > 8 * 1024 * 1024:
        abort(413)
    digest = hashlib.sha256(data).hexdigest()
    path = ROOT / 'bundles' / (digest + '.json')
    if not path.exists():
        # Cache eviction never removes an instance's separate immutable snapshot.
        cached = sorted((ROOT / 'bundles').glob('*.json'), key=lambda p: p.stat().st_mtime)
        used = sum(p.stat().st_size for p in cached)
        while used + len(data) > 128 * 1024 * 1024 and cached:
            old = cached.pop(0)
            used -= old.stat().st_size
            old.unlink()
        if shutil.disk_usage(ROOT).free < 2 * 1024**3:
            abort(507, 'Runtime storage reserve reached')
        temporary = path.with_suffix('.' + uuid.uuid4().hex + '.tmp')
        temporary.write_bytes(data)
        temporary.chmod(0o444)
        os.replace(temporary, path)
    return path, digest


def running(iid):
    response = docker('inspect', '--format', '{{json .State}}', identity(iid), check=False)
    return bool(response.returncode == 0 and json.loads(response.stdout).get('Running'))


def reconcile(iid):
    """Refresh Docker-assigned ports before routing after a daemon/host restart."""
    item = record(iid)
    if not item or item['status'] != 'running':
        return
    response = docker('inspect', identity(iid), check=False)
    if response.returncode:
        return
    actual = json.loads(response.stdout)[0]
    if not actual['State']['Running']:
        return
    started = actual['State']['StartedAt']
    bindings = actual['NetworkSettings']['Ports'].get('8080/tcp')
    port = int(bindings[0]['HostPort']) if bindings else None
    if item['mode'] != 'process' and (not bindings or bindings[0]['HostIp'] != '127.0.0.1'):
        raise RuntimeError('Running program has no private HTTP binding')
    if started == item['container_started'] and port == item['port']:
        return
    if item['mode'] == 'process' and started != item['container_started']:
        docker('exec', identity(iid), 'sh', '-c', "printf 'null\\n' > /proc/1/fd/0")
    with contextlib.closing(connection('state.db')) as db, db:
        db.execute('UPDATE instance SET port=?,started=?,container_started=? WHERE iid=?',
                   (port, datetime.datetime.fromisoformat(started.replace('Z', '+00:00')).timestamp(), started, iid))


def reconcile_all():
    with contextlib.closing(connection('state.db', readonly=True)) as db:
        ids = [row[0] for row in db.execute("SELECT iid FROM instance WHERE status='running'")]
    for iid in ids:
        reconcile(iid)


@app.get('/health')
def health():
    docker('info', '--format', '{{.ServerVersion}}')
    return jsonify(status='ok')


@app.post('/bundles')
def provide():
    data = request.get_json()
    owner(data.get('owner'))
    _, digest = snapshot(data.get('definition'))
    return jsonify(digest=digest)


@app.post('/instances/<iid>/start')
def start(iid):
    name = identity(iid)
    data = request.get_json()
    user = owner(data.get('owner'))
    mode = data.get('mode')
    if user == 'system' or mode not in ('httpPost', 'debug', 'process'):
        abort(400, 'Invalid run mode or owner')
    path, digest = snapshot(data.get('definition'))
    if mode == 'process' and data['definition']['blueprints'][data['definition']['main']].get('services', {}).get('main', {}).get('in', {}).get('type') != 'trigger':
        abort(400, 'Background processes require a trigger input')
    with contextlib.closing(connection('state.db')) as db:
        # Serialize creation and reserve capacity before contacting Docker.
        db.execute('BEGIN IMMEDIATE')
        existing = db.execute('SELECT * FROM instance WHERE iid=?', (iid,)).fetchone()
        if existing and existing['owner'] != user:
            abort(404)
        if existing and running(iid):
            return jsonify(dict(existing))
        if not existing and db.execute('SELECT COUNT(*) FROM instance').fetchone()[0] >= 12:
            abort(429, 'Runtime capacity reached')
        if not existing and db.execute('SELECT COUNT(*) FROM instance WHERE owner=?', (user,)).fetchone()[0] >= 5:
            abort(429, 'Account runtime capacity reached')
        docker('rm', '-f', name, check=False)
        bundle = ROOT / 'instances' / (iid + '.json')
        shutil.copyfile(path, bundle)
        bundle.chmod(0o444)
        lock = ROOT / 'locks' / iid
        lock.touch(exist_ok=True)
        lock.chmod(0o644)
        options = ['create', '--name', name, '--label', 'slang.instance=' + iid,
            '--network', 'slang-programs', '--dns', '1.1.1.1', '--read-only',
            '--cap-drop', 'ALL', '--security-opt', 'no-new-privileges', '--user', '65532:65532',
            '--pids-limit', '64', '--memory', '192m', '--memory-swap', '192m', '--cpus', '.25',
            '--ulimit', 'nofile=128:128', '--tmpfs', '/tmp:rw,noexec,nosuid,nodev,size=32m',
            '--log-driver', 'local', '--log-opt', 'max-size=5m', '--log-opt', 'max-file=2',
            '--restart', 'unless-stopped', '--mount', f'type=bind,src={bundle},dst=/bundle.json,readonly']
        if mode != 'process':
            options.extend(['--publish', '127.0.0.1::8080'])
        else:
            options.append('--interactive')
        cli_mode = 'httpPost' if mode == 'debug' else mode
        options.extend([IMAGE, '-mode', cli_mode, '-bind', '0.0.0.0:8080', '/bundle.json'])
        started = time.time()
        docker(*options)
        docker('start', name)
        container_started = json.loads(docker('inspect', '--format', '{{json .State}}', name).stdout)['StartedAt']
        if mode == 'process':
            # Deliver the initial trigger without closing the container's input
            # pipe; subsequent events can come from schedules/delegates.
            docker('exec', name, 'sh', '-c', "printf 'null\\n' > /proc/1/fd/0")
        port = None
        if mode != 'process':
            port = int(docker('port', name, '8080/tcp').stdout.decode().strip().rsplit(':', 1)[1])
        db.execute('INSERT INTO instance(iid,owner,mode,port,status,public,started,digest,log_cursor) VALUES(?,?,?,?,?,?,?,?,?) '
            'ON CONFLICT(iid) DO UPDATE SET mode=excluded.mode,port=excluded.port,status=excluded.status,public=0,started=excluded.started,digest=excluded.digest,log_cursor=excluded.log_cursor',
            (iid, user, mode, port, 'running', 0, started, digest, str(started)))
        db.execute('UPDATE instance SET container_started=? WHERE iid=?', (container_started, iid))
        db.commit()
    # Give invalid blueprints time to fail before reporting them as started.
    time.sleep(.25)
    if not running(iid):
        docker('update', '--restart=no', name, check=False)
        docker('stop', '--time', '1', name, check=False)
        with contextlib.closing(connection('state.db')) as db, db:
            db.execute("UPDATE instance SET status='failed', public=0 WHERE iid=?", (iid,))
        collect_logs(iid)
        abort(422, 'The engine rejected this blueprint; inspect its logs')
    return jsonify(record(iid))


@app.post('/instances/<iid>/stop')
def stop(iid):
    owned(iid)
    docker('stop', '--time', '3', identity(iid), check=False)
    collect_logs(iid)
    with contextlib.closing(connection('state.db')) as db, db:
        db.execute("UPDATE instance SET status='stopped',public=0 WHERE iid=?", (iid,))
    return jsonify(record(iid))


@app.post('/instances/<iid>/delete')
def delete(iid):
    owned(iid)
    docker('rm', '-f', identity(iid), check=False)
    with contextlib.closing(connection('state.db')) as db, db:
        db.execute('DELETE FROM instance WHERE iid=?', (iid,))
    (ROOT / 'instances' / (iid + '.json')).unlink(missing_ok=True)
    return jsonify(status='deleted')


@app.get('/instances/<iid>')
def status(iid):
    item = owned(iid)
    actual = running(iid)
    if not actual and item['status'] == 'running':
        with contextlib.closing(connection('state.db')) as db, db:
            db.execute("UPDATE instance SET status='failed',public=0 WHERE iid=?", (iid,))
        item['status'] = 'failed'
    return jsonify(dict(item, uptime=int(time.time() - item['started']) if actual else -1))


@app.post('/instances/<iid>/route')
def route(iid):
    item = owned(iid)
    public = item['mode'] == 'httpPost' and item['status'] == 'running' and running(iid)
    with contextlib.closing(connection('state.db')) as db, db:
        db.execute('UPDATE instance SET public=? WHERE iid=?', (int(public), iid))
    return jsonify(public=public)


@app.post('/instances/<iid>/invoke')
def debug(iid):
    item = owned(iid)
    code, data = invoke(item, json.dumps(request.get_json().get('input')).encode())
    return Response(data, status=code, content_type='application/json')


def collect_logs(iid):
    item = record(iid)
    result = docker('logs', '--timestamps', '--tail', '200', '--since', item['log_cursor'], identity(iid), check=False)
    latest = item['log_cursor']
    for line in (result.stdout + result.stderr).decode(errors='replace').splitlines():
        timestamp, _, message = line.partition(' ')
        if not timestamp.endswith('Z') or not message:
            continue
        # Docker's --since boundary includes the last line again. The outbox is
        # emptied after delivery, so its UUID uniqueness alone cannot suppress
        # retransmitting that same line on every telemetry poll.
        if item['log_cursor'].endswith('Z') and timestamp <= item['log_cursor']:
            continue
        try:
            parsed = json.loads(message)
            level = parsed.get('level', 'info')
        except (ValueError, AttributeError):
            level = 'info'
            message = json.dumps({'level': level, 'msg': message[:8000], 'time': timestamp})
        unique = str(uuid.uuid5(uuid.NAMESPACE_URL, iid + ':' + line))
        event('logLineReceived', {'program': 'transit:' + iid, 'level': level, 'time': int(time.time()), 'message': message[:16384]}, unique)
        latest = max(latest, timestamp)
    if latest != item['log_cursor']:
        with contextlib.closing(connection('state.db')) as db, db:
            db.execute('UPDATE instance SET log_cursor=? WHERE iid=?', (latest, iid))


@app.get('/events')
def events():
    reconcile_all()
    with contextlib.closing(connection('state.db', readonly=True)) as db:
        ids = [row[0] for row in db.execute('SELECT iid FROM instance')]
    for iid in ids:
        collect_logs(iid)
    with contextlib.closing(connection('events.db')) as db:
        rows = db.execute('SELECT id,topic,data FROM event ORDER BY rowid LIMIT 200').fetchall()
    for row in rows:
        data = json.loads(row['data'])
        if row['topic'] == 'instanceHttpAccessed' and data.get('status') == 504:
            item = record(data['instance_id'])
            if item and item['status'] == 'running':
                docker('stop', '--time', '1', identity(item['iid']), check=False)
                with contextlib.closing(connection('state.db')) as db, db:
                    db.execute("UPDATE instance SET status='failed',public=0 WHERE iid=?", (item['iid'],))
    return jsonify([{'id': row['id'], 'topic': row['topic'], 'data': json.loads(row['data'])} for row in rows])


@app.post('/events/ack')
def acknowledge():
    ids = request.get_json().get('ids')
    if not isinstance(ids, list) or len(ids) > 200 or any(not isinstance(i, str) for i in ids):
        abort(400)
    with contextlib.closing(connection('events.db')) as db, db:
        db.executemany('DELETE FROM event WHERE id=?', [(i,) for i in ids])
    return jsonify(acknowledged=len(ids))
