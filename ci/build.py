"""Build the daemon and runner release archives; any failed command aborts."""

import os
from pathlib import Path
import re
import subprocess
import sys
import tarfile
import time
import zipfile

ROOT = Path(__file__).resolve().parent.parent
RELEASE_DIR = ROOT / "ci" / "release"
# Go no longer supports darwin/386. Keep the other existing release targets.
TARGETS = [("darwin", "amd64"), ("linux", "386"), ("linux", "amd64"),
           ("windows", "386"), ("windows", "amd64")]


def build(version, password="", output_dir=RELEASE_DIR):
    if not re.fullmatch(r"v[0-9]+\.[0-9]+\.[0-9]+(?:[-+][A-Za-z0-9.-]+)?", version):
        raise ValueError("version must look like v1.2.3 or v1.2.3-rc.1")
    output_dir = Path(output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)
    build_time = str(int(time.time()))
    for command in ("slangd", "slang"):
        for target_os, arch in TARGETS:
            name = f"{command}-{version.replace('.', '_')}-{target_os}-{arch}"
            binary = output_dir / (name + (".exe" if target_os == "windows" else ""))
            env = dict(os.environ, GOOS=target_os, GOARCH=arch, CGO_ENABLED="0")
            args = ["go", "build", "-trimpath"]
            if command == "slangd":
                args += ["-ldflags", f"-X main.Version={version} -X main.BuildTime={build_time}"]
            args += ["-o", str(binary), f"./cmd/{command}"]
            print(f"Building {name}", flush=True)
            subprocess.run(args, cwd=ROOT, env=env, check=True)
            if command == "slangd" and target_os == "windows" and password:
                signed = binary.with_name("signed_" + binary.name)
                # Supply the password through stdin, not command-line arguments.
                subprocess.run([
                    "osslsigncode", "sign", "-pkcs12", str(ROOT / "ci" / "b6k_csc.p12"),
                    "-readpass", "/dev/stdin", "-in", str(binary), "-out", str(signed),
                ], input=password + "\n", text=True, check=True)
                signed.replace(binary)
            if target_os == "windows":
                with zipfile.ZipFile(output_dir / (name + ".zip"), "w", zipfile.ZIP_DEFLATED) as archive:
                    archive.write(binary, binary.name)
            else:
                with tarfile.open(output_dir / (name + ".tar.gz"), "w:gz") as archive:
                    archive.add(binary, arcname=binary.name)
            binary.unlink()


if __name__ == "__main__":
    if len(sys.argv) != 2:
        sys.exit("Usage: python3 ci/build.py v1.2.3")
    build(sys.argv[1], os.environ.get("B6K_CS_PW", ""))
