import collections
import io
import json
import unittest
import uuid
import zipfile
from unittest.mock import patch

import gateway as g


class GatewayTests(unittest.TestCase):
    def setUp(self):
        g.key = b"test key, not used in production"

    def test_cookie_requires_valid_signature_and_canonical_id(self):
        sid = uuid.uuid4().hex
        self.assertEqual(sid, g.verify(g.sign(sid)))
        self.assertIsNone(g.verify(sid + ".forged"))
        self.assertIsNone(g.verify(g.sign("../other-user")))
        self.assertIsNone(g.verify(""))

    def test_catalog_removes_transitive_unavailable_dependencies(self):
        catalog = [
            {"type": "elementary", "def": {"id": "math"}},
            {"type": "library", "def": {"id": "double", "operators": {"x": {"operator": "math"}}}},
            {"type": "library", "def": {"id": "web", "operators": {"x": {"operator": "http"}}}},
            {"type": "library", "def": {"id": "unsafe", "operators": {"x": {"operator": "web"}}}},
            {"type": "local", "def": {"id": "unfinished"}},
        ]
        self.assertEqual(["math", "double", "unfinished"], [o["def"]["id"] for o in g.compatible_catalog(catalog)])

    def test_legacy_save_becomes_bundle_and_rejects_oversized_graph(self):
        bp = {"id": str(uuid.uuid4()), "services": {"main": {}}}
        bundle = g.validate_bundle(bp, [])
        self.assertEqual(bp, bundle["blueprints"][bp["id"]])
        bp["operators"] = {str(i): {} for i in range(101)}
        with self.assertRaises(ValueError):
            g.validate_bundle(bp, [])

    def test_upload_zip_is_read_without_extracting_paths(self):
        out = io.BytesIO()
        with zipfile.ZipFile(out, "w") as z:
            z.writestr("../../workspace.slang", json.dumps({"main": "example"}))
        self.assertEqual({"main": "example"}, g.unpack_upload(out.getvalue(), "application/zip"))

    def test_zip_bomb_is_rejected_before_read(self):
        out = io.BytesIO()
        with zipfile.ZipFile(out, "w", zipfile.ZIP_DEFLATED) as z:
            z.writestr("workspace.slang", "x" * (g.MAX_BODY + 1))
        with self.assertRaises(ValueError):
            g.unpack_upload(out.getvalue(), "application/zip")

    def test_distinct_cookies_get_distinct_workspaces(self):
        a, b = g.Session(uuid.uuid4().hex), g.Session(uuid.uuid4().hex)
        with patch.dict(g.sessions, {a.sid: a, b.sid: b}, clear=True):
            self.assertIs(a, g.get_session(g.sign(a.sid)))
            self.assertIs(b, g.get_session(g.sign(b.sid)))

    def test_output_history_bounds_bytes_and_keeps_recent_results(self):
        history = collections.deque(maxlen=50)
        for i in range(20):
            g.record_output(history, {"sequence": i, "data": "x" * (g.MAX_OUTPUT - 100)})
        self.assertLessEqual(len(json.dumps(list(history)).encode()), g.MAX_HISTORY)
        self.assertEqual(19, history[-1]["sequence"])
        self.assertGreater(len(history), 0)


if __name__ == "__main__":
    unittest.main()
