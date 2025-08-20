<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="logo-monochrome.svg">
    <img alt="Linkleaf CLI" src="logo.svg" width="420">
  </picture>
</p>

<h1 align="center">Linkleaf CLI</h1>

<p align="center">
  Manage <strong>protobuf-only</strong> Linkleaf feeds (<code>linkleaf.v1</code>) with a tiny Go CLI.
</p>

---

## Overview

`linkleaf` reads and writes a single **binary protobuf** file (`.pb`) containing a `linkleaf.v1.Feed`.
There is **no JSON** anywhere—storage and I/O are **protobuf wire format only**.

**Schema:** [`proto/linkleaf/v1/feed.proto`](proto/linkleaf/v1/feed.proto)
**Go module:** `github.com/doriancodes/linkleaf-cli`
**Generated package import:** `github.com/doriancodes/linkleaf-cli/proto/linkleaf/v1`

---

## Quick start

```bash
# Clone and enter
git clone https://github.com/doriancodes/linkleaf-cli.git
cd linkleaf-cli

# 1) Install protoc-gen-go (requires protoc installed on your system)
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
export PATH="$(go env GOPATH)/bin:$PATH"

# 2) Generate Go code from the proto (source-relative output)
protoc -I=proto --go_out=paths=source_relative:. proto/linkleaf/v1/feed.proto

# 3) Build the CLI
go build -o linkleaf ./cmd/linkleaf
```

## Usage

```bash
linkleaf – protobuf-only feed manager (linkleaf.v1)

Usage:
  linkleaf init  <file.pb> [-title "My Feed"] [-version 1]
  linkleaf add   -file <file.pb> -title "..." -url "..." -date YYYY-MM-DD \
                 [-summary "..."] [-tags a,b,c] [-via URL] [-id ID]
  linkleaf list  <file.pb>
  linkleaf print <file.pb>

Notes:
  • Data is stored ONLY in protobuf binary files (.pb).
  • "add" prepends links (newest first). If -id is empty: sha256(url+"|"+date)[:12].
  • "init" creates the file if it doesn't exist; "add" will also create on demand.
```

## Examples

```bash
# Create a new feed (protobuf file)
./linkleaf init feed.pb -title "My Links" -version 1

# Add a link (written directly into feed.pb as protobuf)
./linkleaf add -file feed.pb \
  -title "Protobuf Best Practices" \
  -url "https://developers.google.com/protocol-buffers/docs/best-practices" \
  -date 2025-08-18 \
  -tags protobuf,design

# List links (human-readable output; data stays in protobuf)
./linkleaf list feed.pb

# Pretty text dump for inspection (still reads protobuf)
./linkleaf print feed.pb

```
