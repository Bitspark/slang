"""Shared, bounded runtime routing and telemetry; no Docker access here."""
import contextlib
import fcntl
import json
import os
import re
import sqlite3
import time
import uuid
from pathlib import Path

import requests

ROOT = Path('/var/lib/slang-runtime')
ID = re.compile(r'[0-9a-f]{32}')
MAX_BODY = 2 * 1024 * 1024


def connection(name, readonly=False):
    path = ROOT / 'telemetry' / name if name == 'events.db' else ROOT / name
    if readonly:
        result = sqlite3.connect(f'file:{path}?mode=ro', uri=True, timeout=5)
    else:
        result = sqlite3.connect(path, timeout=5)
    result.row_factory = sqlite3.Row
    return result


def record(iid):
    if not ID.fullmatch(iid):
        raise ValueError('Invalid deployment ID')
    with contextlib.closing(connection('state.db', readonly=True)) as db:
        item = db.execute('SELECT * FROM instance WHERE iid=?', (iid,)).fetchone()
    return dict(item) if item else None


def event(topic, data, message_id=None):
    with contextlib.closing(connection('events.db')) as db, db:
        if db.execute('SELECT COUNT(*) FROM event').fetchone()[0] >= 10000:
            raise RuntimeError('Telemetry queue is full')
        db.execute('INSERT OR IGNORE INTO event(id,topic,data) VALUES(?,?,?)',
                   (message_id or str(uuid.uuid4()), topic, json.dumps(data)))


def invoke(item, body):
    if not item or item['status'] != 'running' or not item['port']:
        return 404, b'{"message":"Deployment is not running"}'
    if len(body) > MAX_BODY:
        return 413, b'{"message":"Input exceeds 2 MiB"}'
    lock_path = ROOT / 'locks' / item['iid']
    # The control agent pre-creates locks. The router never follows user paths.
    with lock_path.open('r') as lock:
        try:
            fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError:
            return 429, b'{"message":"Deployment is busy"}'
        try:
            with requests.post(f"http://127.0.0.1:{item['port']}/", data=body,
                headers={'Content-Type': 'application/json'}, timeout=(2, 15), stream=True) as response:
                output = bytearray()
                for part in response.iter_content(16384):
                    output.extend(part)
                    if len(output) > MAX_BODY:
                        return 502, b'{"message":"Output exceeds 2 MiB"}'
                status, result = response.status_code, bytes(output)
        except requests.RequestException:
            status, result = 504, b'{"message":"Program failed or timed out"}'
        finally:
            fcntl.flock(lock, fcntl.LOCK_UN)
    event('instanceHttpAccessed', {'instance_id': item['iid'], 'owner': item['owner'], 'status': status})
    if status >= 400:
        event('logLineReceived', {'program': 'transit:' + item['iid'], 'level': 'error',
            'time': int(time.time()), 'message': json.dumps({'level': 'error', 'msg': result.decode(errors='replace')[:1000], 'time': time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime())})})
    return status, result
