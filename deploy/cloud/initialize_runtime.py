"""Initialize runtime state without resetting deployed instances or telemetry."""
import grp
import os
from pathlib import Path
import sqlite3

root = Path('/var/lib/slang-runtime')
group = grp.getgrnam('slang-router').gr_gid
root.mkdir(exist_ok=True, mode=0o755)
for child in ('instances', 'bundles', 'locks'):
    (root / child).mkdir(exist_ok=True, mode=0o755)
with sqlite3.connect(root / 'state.db') as db:
    db.execute('CREATE TABLE IF NOT EXISTS instance(iid TEXT PRIMARY KEY, owner TEXT NOT NULL, mode TEXT NOT NULL, port INTEGER, status TEXT NOT NULL, public INTEGER NOT NULL DEFAULT 0, started REAL NOT NULL, digest TEXT NOT NULL, log_cursor TEXT NOT NULL)')
    if 'container_started' not in {row[1] for row in db.execute('PRAGMA table_info(instance)')}:
        db.execute("ALTER TABLE instance ADD COLUMN container_started TEXT NOT NULL DEFAULT ''")
os.chmod(root / 'state.db', 0o644)
# A separate writable directory lets the unprivileged router create SQLite journals.
events = root / 'telemetry'
events.mkdir(exist_ok=True, mode=0o770)
os.chown(events, 0, group)
events.chmod(0o770)
with sqlite3.connect(events / 'events.db') as db:
    db.execute('CREATE TABLE IF NOT EXISTS event(id TEXT PRIMARY KEY, topic TEXT NOT NULL, data TEXT NOT NULL)')
    db.execute('CREATE TABLE IF NOT EXISTS invocation_rate(iid TEXT PRIMARY KEY, window INTEGER NOT NULL, count INTEGER NOT NULL)')
os.chown(events / 'events.db', 0, group)
os.chmod(events / 'events.db', 0o660)
print('Runtime state initialized without resetting data')
