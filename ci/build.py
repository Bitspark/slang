import sys
from os import system
from time import gmtime, strftime
from utils import execute_commands

OS = ['darwin', 'linux', 'windows']
ARCHS = ['386', 'amd64']

if __name__ == '__main__':
    if len(sys.argv) != 2:
        print('Usage: python3 build.py vx.y.z')
        exit(-1)

    version = sys.argv[1]
    versioned_dist = 'slangd-' + version.replace('.', '_')
    build_time = strftime("%Y-%m-%d %H:%M:%S", gmtime())

    ldflags = f"-X main.Version=`{version}` "
    ldflags += f"-X main.BuildTime=`{build_time}` "

    for os in OS:
        for arch in ARCHS:
            filename_with_ending = filename = f"{versioned_dist}-{os}-{arch}"
            if os == 'windows':
                filename_with_ending += ".exe"
                compress_cmd = f"zip {filename}.zip {filename_with_ending}"
            else:
                compress_cmd = f"tar -czvf {filename}.tar.gz {filename_with_ending}"

            execute_commands([
                f"env GOOS={os} GOARCH={arch} go build -ldflags \"{ldflags}\" -o ./ci/release/{filename_with_ending} ./cmd/slangd",
            ])

            os.chdir("./ci/release/")
            execute_commands([
                compress_cmd,
                "rm {filename_with_ending}",
            ])
            os.chdir("../..")
