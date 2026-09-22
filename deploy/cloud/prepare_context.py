"""Build an explicit source context without historical credentials/config files.

Run on the controller with sibling component checkouts. The resulting directory
is disposable build input, never a copy of .env/PEM/Now deployment secrets.
"""
import argparse
import json
import shutil
import subprocess
from pathlib import Path


def copy_sources(source, target, suffixes, excluded=()):
    for item in source.rglob('*'):
        rel = item.relative_to(source)
        if any(part.startswith('.') or part in ('node_modules', '__pycache__', 'tests', 'migrations') for part in rel.parts):
            continue
        if item.is_file() and item.suffix in suffixes and item.name not in excluded:
            out = target / rel
            out.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(item, out)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--components', required=True, help='JSON mapping service name to source directory')
    parser.add_argument('--output', required=True)
    options = parser.parse_args()
    output = Path(options.output)
    output.mkdir(parents=True, exist_ok=True)
    manifest = {}
    for name, directory in json.loads(Path(options.components).read_text()).items():
        source = Path(directory)
        dest = output / ('telstar' if name == 'telstar' else 'services/' + name)
        # A new empty context prevents removed/excluded files surviving a rebuild.
        if dest.exists():
            raise SystemExit(f'Use a fresh output directory: {dest} already exists')
        copy_sources(source, dest, {'.py', '.toml', '.png', '.html', '.tpl', '.js', '.css', '.json'},
                     excluded=('now.json', 'Pipfile.lock', 'migrations.json'))
        for filename in ('LICENSE',):
            if (source / filename).exists():
                shutil.copy2(source / filename, dest / filename)
        manifest[name] = dict(revision=subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=source, text=True).strip(),
                              dirty=bool(subprocess.check_output(['git', 'status', '--porcelain'], cwd=source, text=True).strip()))
    shutil.copytree(Path(__file__).parent, output / 'cloud', ignore=shutil.ignore_patterns('__pycache__'))
    runtime = Path(__file__).resolve().parents[2]
    manifest['slang'] = dict(revision=subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=runtime, text=True).strip(),
                             dirty=bool(subprocess.check_output(['git', 'status', '--porcelain'], cwd=runtime, text=True).strip()))
    (output / 'source-manifest.json').write_text(json.dumps(manifest, indent=2) + '\n')
    print('Prepared service sources; no environment files or PEM keys copied')


if __name__ == '__main__':
    main()
