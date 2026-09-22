"""Fetch immutable, checksum-verified releases; no Node toolchain is required."""
import hashlib
import io
import json
from pathlib import Path
import re
import shutil
import sys
import urllib.request
import zipfile

RELEASES = {
    "ui": ("slang-ui", "v0.2.5a", "slang-ui-v0_2_5a.zip", "f0f0c19e315c93eb8415430d038e88419fc7d70e80ee0833cf3803d822674dd6"),
    "lib": ("slang-lib", "v0.1.9", "slang-lib-v0_1_9.zip", "a3728b8c447329bff170158924be1e8ba9ba899fa15a3a6f88156ad72c7b32e8"),
}

SOURCE = Path(__file__).parent


def install_website(root, archive_path=None):
    """Install the independently versioned website, with verified dependencies."""
    lock = json.loads((SOURCE / "website.lock.json").read_text())
    if archive_path:
        data = Path(archive_path).read_bytes()
    else:
        with urllib.request.urlopen(lock["url"], timeout=60) as response:
            data = response.read()
    if hashlib.sha256(data).hexdigest() != lock["sha256"]:
        raise RuntimeError("Release checksum mismatch: slang-website")
    dest = root / "assets/website" / lock["version"]
    with zipfile.ZipFile(io.BytesIO(data)) as archive:
        for item in archive.infolist():
            if item.is_dir():
                continue
            relative = Path(item.filename)
            if relative.is_absolute() or ".." in relative.parts or "\\" in item.filename or ":" in item.filename:
                raise RuntimeError("Invalid website asset path")
            if relative.parts[0] not in {"site", "editor-head.html", "editor-shell.html", "LICENSE", "NOTICE"}:
                raise RuntimeError("Unexpected website asset: " + item.filename)
        for item in archive.infolist():
            if not item.is_dir():
                target = dest / item.filename
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_bytes(archive.read(item))
    if not (dest / "site/index.html").is_file():
        raise RuntimeError("Website release is missing its index")
    for name in ["editor-head.html", "editor-shell.html"]:
        if not (dest / name).is_file():
            raise RuntimeError("Website release is missing " + name)
    shutil.copytree(dest / "site", root / "site", dirs_exist_ok=True)
    (root / "site/website-source.json").write_text(json.dumps(lock, indent=2) + "\n", encoding="utf-8")
    print("Verified and installed slang-website", lock["version"])
    return dest


def style_editor(html, website):
    """Add the released website shell; leave Angular's runtime scripts intact."""
    html = re.sub(r'<link[^>]+rel="icon"[^>]*>', '', html)
    html = re.sub(r'<style>#sl-feedback-box.*?</style>', '', html, flags=re.S)
    html = re.sub(r'<div id="restore-bar">.*?</div>', '', html, flags=re.S)
    html = re.sub(r'<!-- slang-design:start -->.*?<!-- slang-design:end -->', '', html, flags=re.S)
    html = re.sub(r'<body[^>]*>', '<body id="slang-playground" class="sd-scope slang-editor">', html, count=1)
    html = html.replace('<title>Slang</title>', '<title>Slang playground</title>')
    head = (website / "editor-head.html").read_text(encoding="utf-8")
    shell = (website / "editor-shell.html").read_text(encoding="utf-8")
    wrap = lambda text: '<!-- slang-design:start -->' + text + '<!-- slang-design:end -->'
    html = html.replace('</head>', wrap(head) + '</head>')
    html = html.replace('<app-root>', wrap(shell) + '<app-root>')
    return html


def install_frontend(root, archive_path=None):
    index = root / "assets/ui/index.html"
    if not index.is_file():
        raise RuntimeError("Install the released UI before updating the frontend")
    website = install_website(root, archive_path)
    index.write_text(style_editor(index.read_text(encoding="utf-8"), website), encoding="utf-8")
    # The released guide iframe hard-codes production. Keep it same-origin,
    # including previews and independently hosted deployments.
    for bundle in (root / "assets/ui").glob("main.*.js"):
        text = bundle.read_text(encoding="utf-8")
        text = text.replace("https://tryslang.com/slang-app/index-1", "/slang-app/index-1/")
        bundle.write_text(text, encoding="utf-8")


def main():
    root = Path(sys.argv[1])
    archive_path = None
    if "--website-archive" in sys.argv:
        archive_path = sys.argv[sys.argv.index("--website-archive") + 1]
    if "--frontend-only" in sys.argv[2:]:
        install_frontend(root, archive_path)
        return
    for component, (repo, tag, filename, checksum) in RELEASES.items():
        url = f"https://github.com/Bitspark/{repo}/releases/download/{tag}/{filename}"
        with urllib.request.urlopen(url, timeout=30) as response:
            data = response.read()
        if hashlib.sha256(data).hexdigest() != checksum:
            raise RuntimeError("Release checksum mismatch: " + component)
        dest = root / "assets" / component
        dest.mkdir(parents=True, exist_ok=True)
        with zipfile.ZipFile(io.BytesIO(data)) as archive:
            for item in archive.infolist():
                parts = Path(item.filename).parts
                if item.is_dir() or ".." in parts:
                    continue
                if component == "ui":
                    relative = Path(*parts[1:])
                else:
                    # Only actual blueprints; exclude old _package.yaml metadata.
                    if not item.filename.endswith((".yaml", ".json")) or Path(item.filename).name.startswith("_"):
                        continue
                    relative = Path(item.filename).name
                target = dest / relative
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_bytes(archive.read(item))
        print("Verified and installed", component, tag)
    install_frontend(root, archive_path)
    shutil.copytree(SOURCE / "examples", root / "examples", dirs_exist_ok=True)


if __name__ == "__main__":
    main()
