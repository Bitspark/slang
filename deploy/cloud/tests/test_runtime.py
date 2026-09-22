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


class RuntimeReboot(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.patch = patch.object(common, 'ROOT', self.root)
        self.patch.start()
        self.addCleanup(self.patch.stop)
        self.iid = 'b' * 32
        with sqlite3.connect(self.root / 'state.db') as db:
            db.execute('CREATE TABLE instance(iid TEXT PRIMARY KEY, mode TEXT, port INTEGER, status TEXT, public INTEGER, started REAL, container_started TEXT)')
            db.execute('INSERT INTO instance VALUES(?,?,?,?,?,?,?)', (self.iid, 'httpPost', 32000, 'running', 1, 1.0, 'old'))

    def inspection(self, mode='httpPost'):
        return SimpleNamespace(returncode=0, stdout=json.dumps([{
            'State': {'Running': True, 'StartedAt': '2026-09-22T05:00:00.123456789Z'},
            'NetworkSettings': {'Ports': {'8080/tcp': [{'HostIp': '127.0.0.1', 'HostPort': '32770'}]} if mode != 'process' else {}}
        }]).encode())

    def test_reboot_updates_route_before_invocation_without_losing_public_flag(self):
        with patch.object(agent, 'docker', return_value=self.inspection()):
            agent.reconcile_all()
        item = common.record(self.iid)
        item['owner'] = 'c' * 32
        self.assertEqual(item['port'], 32770)
        self.assertEqual(item['public'], 1)
        with patch.object(common, 'requests') as requests, patch.object(common, 'event'):
            (self.root / 'locks').mkdir()
            (self.root / 'locks' / self.iid).touch()
            response = requests.post.return_value.__enter__.return_value
            response.iter_content.return_value = [b'"echo"']
            response.status_code = 200
            self.assertEqual(common.invoke(item, b'"echo"'), (200, b'"echo"'))
            self.assertEqual(requests.post.call_args.args[0], 'http://127.0.0.1:32770/')

    def test_restarted_background_process_receives_one_initial_trigger(self):
        with sqlite3.connect(self.root / 'state.db') as db:
            db.execute("UPDATE instance SET mode='process',port=NULL,public=0")
        with patch.object(agent, 'docker', return_value=self.inspection('process')) as docker:
            agent.reconcile_all()
            agent.reconcile_all()
            self.assertEqual(sum(call.args[0] == 'exec' for call in docker.call_args_list), 1)

    def test_stopped_program_is_not_restarted_or_made_public(self):
        with sqlite3.connect(self.root / 'state.db') as db:
            db.execute("UPDATE instance SET status='stopped',public=0")
        with patch.object(agent, 'docker') as docker:
            agent.reconcile_all()
            docker.assert_not_called()
        self.assertEqual(common.record(self.iid)['status'], 'stopped')


if __name__ == '__main__':
    unittest.main()
