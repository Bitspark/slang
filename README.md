[![CircleCI](https://circleci.com/gh/Bitspark/slang/tree/master.svg?style=svg&circle-token=ba892aab7dad71da5e2c426eff2a336974d96df0)](https://circleci.com/gh/Bitspark/slang/tree/master)[![codecov](https://codecov.io/gh/Bitspark/slang/branch/master/graph/badge.svg)](https://codecov.io/gh/Bitspark/slang)



# Slang Daemon
<p align="center">
  <img src="https://raw.githubusercontent.com/Bitspark/slang/master/logo.png" height="130">
</p>

*Powered by [Bitspark](https://bitspark.de)*

## About Slang

Slang is a visual flow-based programming language and programming system. It consists of the YAML-based Slang exchange format, the **Slang daemon** and the [Slang UI](https://github.com/Bitspark/slang-ui).

## About Slang daemon

Slang daemon is the service which serves the user web interface (Slang UI) and runs all your operators.
You don't need anything else to start working with Slang, so this here is the place to start.

## How to install

If you want to run Slang, you can simply download the [latest release](https://github.com/Bitspark/slang/releases/latest), unpack and run it. We have binaries for Windows, Linux and MacOS.

### Compile it yourself

If you rather want to compile it yourself, you first need to install [Go](https://golang.org/).

After you have set up Go and cloned the repository, switch to the root directory and run

`go build ./...`

This will fetch all the dependencies. After that, run

`go build -o slangd ./cmd/slangd` (on Windows: `go build -o slangd.exe ./cmd/slangd`)

Alternatly you just can run the daemon without compiling

`go run ./cmd/slangd`

That's it! Now you just need to run `slangd` (on Windows: `slangd.exe`) and Slang will take care of the rest such as downloading the UI and standard library.

### Working on stdlib

If you want to edit stdlib, just run 

`SLANG_LIB=/tmp/ SLANG_DIR=<path-to-slang-lib>/slang go run ./cmd/slangd --only-daemon`

## About Slang runner

Slang runner allows you to run a blueprint via cli. Input data must be passed through STDIN. Output will be written to STDOUT. Input and output data must be valid newline-delimited json (ndjson). As an argument you have to provide a slang file.

`cat data.ndjson | go run ./cmd/slang d37961fd-7677-4b32-ba1c-cbae2bb85999.slang | less`

## Links

- [TrySlang website](http://tryslang.com)
- [Slang UI repository](https://github.com/Bitspark/slang-ui)
- [Bitspark website](https://bitspark.de)
