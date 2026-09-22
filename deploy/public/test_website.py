"""Website release integration, without network access or a running engine."""
import hashlib
import io
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import zipfile

import prepare


class WebsiteInstallerTests(unittest.TestCase):
    def test_verified_release_and_idempotent_editor_shell(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            source = root / "source"
            source.mkdir()
            data = io.BytesIO()
            with zipfile.ZipFile(data, "w") as archive:
                for name, content in {
                    "site/index.html": "Website",
                    "site/design/0.2.1/styles/index.css": "/* shared styles */",
                    "editor-head.html": '<link rel="stylesheet" href="/design/0.2.1/styles/index.css">',
                    "editor-shell.html": '<header>Slang</header>',
                }.items():
                    archive.writestr(name, content)
            payload = data.getvalue()
            artifact = root / "website.zip"
            artifact.write_bytes(payload)
            (source / "website.lock.json").write_text(json.dumps({"version": "0.1.0", "sha256": hashlib.sha256(payload).hexdigest()}))
            ui = root / "assets/ui"
            ui.mkdir(parents=True)
            original = '<html><head><title>Slang</title><link rel="icon" href="favicon.ico"></head><body><app-root></app-root><script src="main.js"></script></body></html>'
            (ui / "index.html").write_text(original)
            (ui / "main.test.js").write_text('url="https://tryslang.com/slang-app/index-1";')
            state = root / "state"
            state.mkdir()
            (state / "visitor.json").write_text('private data')
            with patch.object(prepare, "SOURCE", source):
                prepare.install_frontend(root, artifact)
                first = (ui / "index.html").read_text()
                prepare.install_frontend(root, artifact)
            self.assertEqual(first, (ui / "index.html").read_text())
            self.assertEqual(first.count('<header>'), 1)
            self.assertIn('<script src="main.js"></script>', first)
            self.assertIn('class="sd-scope slang-editor"', first)
            self.assertEqual((ui / "main.test.js").read_text(), 'url="/slang-app/index-1/";')
            self.assertEqual((root / "site/index.html").read_text(), "Website")
            self.assertEqual((state / "visitor.json").read_text(), 'private data')

    def test_bad_checksum_leaves_live_site_untouched(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            (root / "website.lock.json").write_text(json.dumps({"version": "0.1.0", "sha256": "wrong"}))
            (root / "wrong.zip").write_bytes(b"wrong archive")
            (root / "site").mkdir()
            (root / "site/index.html").write_text("Current website")
            with patch.object(prepare, "SOURCE", root), self.assertRaisesRegex(RuntimeError, "checksum mismatch"):
                prepare.install_website(root, root / "wrong.zip")
            self.assertEqual((root / "site/index.html").read_text(), "Current website")


if __name__ == "__main__":
    unittest.main()
