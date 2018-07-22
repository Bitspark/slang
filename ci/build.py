import sys
from os import system
from utils import execute_commands

OS = ['darwin', 'linux', 'windows']
ARCHS = ['386', 'amd64']

if __name__ == '__main__':
    if len(sys.argv) != 2:
        print('Usage: python3 build.py vx.y.z')
        exit(-1)

    versioned_dist = 'slangd-' + sys.argv[1].replace('.', '_')

    for os in OS:
        for arch in ARCHS:
            execute_commands([
                f"env GOOS={os} GOARCH={arch} go build -o ./ci/release/{versioned_dist}-{os}-{arch}{'.exe' if os == 'windows' else ''} ./cmd/slangd",
            ])