import json
from pathlib import Path
import sqlite3
import sys
import tempfile
from types import SimpleNamespace
import unittest
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
import runtime_agent as agent
import runtime_common as common


class RuntimeLogs(unittest.TestCase):
    def test_inclusive_docker_cursor_does_not_republish_delivered_line(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            with sqlite3.connect(root / 'state.db') as db:
                db.execute('CREATE TABLE instance(iid TEXT PRIMARY KEY, log_cursor TEXT)')
                db.execute('INSERT INTO instance VALUES(?,?)', ('a' * 32, '1.0'))
            output = SimpleNamespace(stdout=b'2026-09-22T04:00:00.123456789Z null\n', stderr=b'')
            with patch.object(common, 'ROOT', root), patch.object(agent, 'docker', return_value=output), patch.object(agent, 'event') as event:
                agent.collect_logs('a' * 32)
                self.assertEqual(event.call_count, 1)
                self.assertEqual(json.loads(event.call_args.args[1]['message'])['msg'], 'null')
                agent.collect_logs('a' * 32)
                self.assertEqual(event.call_count, 1)


if __name__ == '__main__':
    unittest.main()
