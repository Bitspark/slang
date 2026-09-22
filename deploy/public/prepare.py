"""Fetch immutable, checksum-verified releases; no Node toolchain is required."""
import hashlib
import io
from pathlib import Path
import shutil
import sys
import urllib.request
import zipfile

RELEASES = {
    "ui": ("slang-ui", "v0.2.5a", "slang-ui-v0_2_5a.zip", "f0f0c19e315c93eb8415430d038e88419fc7d70e80ee0833cf3803d822674dd6"),
    "lib": ("slang-lib", "v0.1.9", "slang-lib-v0_1_9.zip", "a3728b8c447329bff170158924be1e8ba9ba899fa15a3a6f88156ad72c7b32e8"),
}


def main():
    root = Path(sys.argv[1])
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
    index = root / "assets/ui/index.html"
    html = index.read_text()
    html = html.replace('<link rel="icon" type="image/x-icon" href="favicon.ico">',
                        '<link rel="icon" type="image/svg+xml" href="/brand/slang-mark.svg">')
    html = html.replace("</head>", '<style>#sl-feedback-box{display:none!important} #restore-bar{position:fixed;bottom:0;left:0;right:0;z-index:9999;padding:5px 12px;background:#101d32;color:#fff;font:12px system-ui} #restore-bar a{color:#8cdeef;margin-right:16px}</style></head>')
    html = html.replace("</body>", '<div id="restore-bar"><a href="/">TrySlang home</a><a href="/workspace/export">Export workspace</a>Private workspace · saved for 30 days of inactivity · export a backup</div></body>')
    index.write_text(html)
    source = Path(__file__).parent
    shutil.copytree(source / "site", root / "site", dirs_exist_ok=True)
    shutil.copytree(source / "examples", root / "examples", dirs_exist_ok=True)


if __name__ == "__main__":
    main()
