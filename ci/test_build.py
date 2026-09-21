import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch
import tarfile
import zipfile

import build


class ReleaseTests(unittest.TestCase):
    def fake_build(self, args, **kwargs):
        self.assertEqual("go", args[0])
        self.assertTrue(kwargs["check"])
        Path(args[args.index("-o") + 1]).write_bytes(b"test executable")

    def test_every_supported_target_is_archived(self):
        with tempfile.TemporaryDirectory() as directory:
            with patch("build.subprocess.run", side_effect=self.fake_build) as run:
                build.build("v1.2.3", output_dir=directory)
            self.assertEqual(10, run.call_count)
            archives = list(Path(directory).iterdir())
            self.assertEqual(10, len(archives))
            self.assertEqual(4, sum(path.suffix == ".zip" for path in archives))
            for path in archives:
                if path.suffix == ".zip":
                    with zipfile.ZipFile(path) as archive:
                        self.assertEqual([path.stem + ".exe"], archive.namelist())
                else:
                    with tarfile.open(path) as archive:
                        self.assertEqual([path.name.removesuffix(".tar.gz")], archive.getnames())

    def test_failed_build_stops_without_publishing_an_archive(self):
        with tempfile.TemporaryDirectory() as directory:
            with patch("build.subprocess.run", side_effect=subprocess.CalledProcessError(1, "go")) as run:
                with self.assertRaises(subprocess.CalledProcessError):
                    build.build("v1.2.3", output_dir=directory)
            self.assertEqual(1, run.call_count)
            self.assertEqual([], list(Path(directory).iterdir()))

    def test_invalid_version_is_rejected_before_running_commands(self):
        with patch("build.subprocess.run") as run:
            with self.assertRaises(ValueError):
                build.build("../../invalid")
            run.assert_not_called()


if __name__ == "__main__":
    unittest.main()
