# Continuous integration and releases

GitHub Actions runs `go vet`, the complete Go test suite with coverage, and
`go build` on Linux, Windows, and macOS for pull requests and pushes to master.
Coverage files are retained as workflow artifacts. The package job also checks
the release tooling and builds all release archives on each pull request.

Pushing a `v*` version tag runs the same checks and publishes a GitHub release
only after all tests and packaging succeed. Publishing uses the built-in
`GITHUB_TOKEN`; it does not require the old CircleCI token or checkout key.
PR jobs have read-only repository permissions.

Archives retain the existing `slang[d]-vX_Y_Z-OS-ARCH` naming. Supported targets
are macOS amd64, Linux 386/amd64, and Windows 386/amd64. Go no longer supports
the previous macOS 386 target. Windows ZIPs contain an `.exe`; Linux and macOS
archives use `.tar.gz`.

To build locally, run `python3 ci/build.py v1.2.3`. Failed builds abort rather
than silently publishing an incomplete set of archives. The optional GitHub
Actions secret `B6K_CS_PW` enables signing of Windows daemon executables with
the existing certificate in `ci/b6k_csc.p12`; without it, Windows executables
are unsigned. This secret must be configured in GitHub separately from CircleCI.
It is passed to packaging only for version tags. Local signing also requires
`osslsigncode` and the `B6K_CS_PW` environment variable.

The CircleCI configuration and badges have been removed. The old CircleCI
project can be unfollowed in CircleCI after this migration is merged.
