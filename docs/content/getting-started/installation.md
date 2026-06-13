---
title: "Installation"
description: "Install dict from a release, with go install, or from source."
weight: 20
---

## Prebuilt binaries

Every [release](https://github.com/tamnd/dictionary-cli/releases) carries archives for Linux, macOS,
and Windows on amd64 and arm64, plus deb, rpm, and apk packages for Linux.
Download, unpack, put `dict` on your `PATH`, done. The `checksums.txt`
on each release is signed with keyless [cosign](https://docs.sigstore.dev/) if
you want to verify before running.

## With Go

```bash
go install github.com/tamnd/dictionary-cli/cmd/dict@latest
```

That puts `dict` in `$(go env GOPATH)/bin`, which is `~/go/bin` unless
you moved it. Make sure that directory is on your `PATH`.

## From source

```bash
git clone https://github.com/tamnd/dictionary-cli
cd dictionary-cli
make build        # produces ./bin/dict
./bin/dict version
```

## Container image

```bash
docker run --rm ghcr.io/tamnd/dict:latest --help
```

## Checking the install

```bash
dict version
```

prints the version and exits.
