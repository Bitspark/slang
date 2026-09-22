# Slang

[![CI](https://github.com/Bitspark/slang/actions/workflows/ci.yml/badge.svg)](https://github.com/Bitspark/slang/actions/workflows/ci.yml)

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/Bitspark/slang-design/a16912ee2938ad9202380c88ce486adf893e5ccf/assets/logo/slang-logo-dark.svg">
    <img src="https://raw.githubusercontent.com/Bitspark/slang-design/a16912ee2938ad9202380c88ce486adf893e5ccf/assets/logo/slang-logo-light.svg" alt="Slang" width="280">
  </picture>
</p>

Slang is a visual, flow-based programming language. Programs are graphs of
connected operators, called **blueprints**, that transform data as it flows
between their inputs and outputs.

This repository contains the Go runtime and two commands:

| Command | Purpose |
| --- | --- |
| `slangd` | Local daemon that manages blueprints, runs operators, and serves the visual editor. |
| `slang` | Standalone runner for JSON blueprint bundles, with pipe and HTTP modes. |

The [visual editor](https://github.com/Bitspark/slang-ui) and
[standard library](https://github.com/Bitspark/slang-lib) live in separate repositories.

## Try it in your browser

Open [Slang Run](https://slang.run/) and sign up or log in. Keep the recovery
code shown at signup. Create or clone a blueprint, edit and save it, then use
the studio's test action to execute it with the Go runner. HTTP deployments
receive a public address under `<deployment-id>.slangapps.com`.

The [product website](https://slang.bitspark.com/) is maintained in
[Bitspark/slang.bitspark.com](https://github.com/Bitspark/slang.bitspark.com).
`tryslang.com` and its former playground aliases now redirect there.

Hosting implementation and operations live in the private
[slang-infra repository](https://github.com/Bitspark/slang-infra). Its
[hosted operations guide](https://github.com/Bitspark/slang-infra/blob/main/deploy/cloud/README.md) covers the original
backend services, Telstar/Meta, the two Docker hosts, isolation, backups and
recovery. The [previous playground guide](https://github.com/Bitspark/slang-infra/blob/main/deploy/public/README.md) retains the
standalone configuration deployed earlier on 21–22 September 2026.

## Run locally from source

Install [Go](https://go.dev/doc/install) and Git, then run these commands in a
terminal, including PowerShell on Windows:

```sh
git clone https://github.com/Bitspark/slang.git
cd slang
go mod download
go run ./cmd/slangd --only-daemon
```

The daemon listens on [http://localhost:5149/](http://localhost:5149/). On first
launch it downloads the published UI and standard library from GitHub, so an
internet connection is required. `--only-daemon` prevents the browser from opening
automatically; omit it to open the UI on startup. Keep the terminal open and use
**Ctrl+C** to stop the daemon.

**UI compatibility:** The last published UI uses older save/run endpoints than
the current source API. Serving its files locally does not make those operations
compatible. Use the hosted playground for the visual editor; the source commands
here are useful for working on the runtime and its current API.

The exact Go version used to validate Linux, Windows, and macOS builds is set in
the [CI workflow](.github/workflows/ci.yml).

### Build executables

From the repository root, on Linux or macOS:

```sh
go build -o slangd ./cmd/slangd
go build -o slang ./cmd/slang
./slangd --only-daemon
```

On Windows, use PowerShell:

```powershell
go build -o slangd.exe ./cmd/slangd
go build -o slang.exe ./cmd/slang
.\slangd.exe --only-daemon
```

### Prebuilt downloads

Check the assets attached to a [GitHub release](https://github.com/Bitspark/slang/releases)
for your operating system and architecture. Download `slangd` for the daemon or
`slang` for the standalone runner, extract the archive, and run the executable
from a terminal.

The historical [v0.1.26 release](https://github.com/Bitspark/slang/releases/tag/v0.1.26)
contains Linux and Intel macOS archives, but **no Windows binaries**. These
archives also predate changes in this repository. Windows users should build
from source using the commands above. See the [release guide](ci/README.md) for
the platforms packaged by the current workflow.

## Run a blueprint without the editor

The `slang` runner expects a **JSON bundle** containing a `main` blueprint UUID
and a `blueprints` map with the blueprint and its dependencies. A single YAML
blueprint or a workspace ZIP is not a runner bundle.

For a minimal example, save this as `echo.slang.json` in the repository root:

```json
{
  "main": "d1e020d6-4414-42e0-a5c5-dd6fda91754e",
  "blueprints": {
    "d1e020d6-4414-42e0-a5c5-dd6fda91754e": {
      "id": "d1e020d6-4414-42e0-a5c5-dd6fda91754e",
      "meta": {"name": "Echo a number"},
      "services": {
        "main": {"in": {"type": "number"}, "out": {"type": "number"}}
      },
      "connections": {"(": [")"]}
    }
  }
}
```

Start an HTTP runner:

```sh
go run ./cmd/slang -mode httpPost -bind localhost:8080 echo.slang.json
```

In a second terminal, send a JSON number. On Linux or macOS:

```sh
curl -H 'Content-Type: application/json' -d '21' http://localhost:8080/
```

On Windows PowerShell:

```powershell
Invoke-RestMethod -Method Post -Uri http://localhost:8080/ -ContentType application/json -Body '21'
```

The response is `21`. Use **Ctrl+C** in the runner's terminal to stop it.

The default `process` mode instead reads newline-delimited JSON from a pipe on
standard input and writes JSON to standard output. Diagnostics go to standard
error. Input values must match the blueprint's input type.

## Daemon configuration

Set these environment variables before starting `slangd`. Paths below are
relative to your user home directory (`~` on Unix, `%USERPROFILE%` on Windows).

| Variable | Default | Purpose |
| --- | --- | --- |
| `SLANG_DIR` | `slang/blueprints` | Writable workspace for your blueprints. |
| `SLANG_LIB_REPO_PATH` | `slang/shared` | Destination for downloaded standard-library releases. |
| `SLANG_LIB` | `slang/shared/slang` | Read-only directory of library blueprints. |
| `SLANG_UI` | `slang/ui` | Directory containing the built UI assets. |

Useful flags:

| Flag | Effect |
| --- | --- |
| `--only-daemon` | Start without opening a browser. |
| `--skip-checks` | Skip downloading or updating the UI and standard library; use already installed or manually supplied files. |
| `--without-ui` | Do not serve the UI. Combine with `--skip-checks` to avoid downloading it. |

Run `go run ./cmd/slangd -help` for all available flags.

To work on a local standard-library checkout, set `SLANG_LIB` to its `slang`
directory and use `--skip-checks` to prevent automatic component updates. To
make that checkout writable through the daemon, use it as `SLANG_DIR` instead
and point `SLANG_LIB` at a separate, empty directory. Keep `SLANG_UI` pointing
at built UI assets if you want the daemon to serve them.

## Development

Run these checks from the repository root:

```sh
go mod verify
go vet ./...
go test -timeout 3m ./...
go build ./...
```

[GitHub Actions](https://github.com/Bitspark/slang/actions/workflows/ci.yml) runs
vet, tests with coverage, and builds on Linux, Windows, and macOS. It also tests
release packaging. Coverage files and release archives are available as workflow
artifacts. See [CI and releases](ci/README.md) for tagged release publishing.

Bug reports and contributions are welcome. Please include your operating system,
`go version` or release version, the command you ran, and a minimal blueprint or
bundle when [filing an issue](https://github.com/Bitspark/slang/issues).

## Related projects

- [Slang Run studio](https://slang.run/)
- [Slang product website](https://slang.bitspark.com/)
- [Slang UI](https://github.com/Bitspark/slang-ui)
- [Slang standard library](https://github.com/Bitspark/slang-lib)
- [Slang examples](https://github.com/Bitspark/slang-examples)

## License

Slang is licensed under the [Apache License 2.0](LICENSE).
