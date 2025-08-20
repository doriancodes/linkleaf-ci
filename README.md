<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="logo-monochrome.svg">
    <img alt="Linkleaf CLI" src="logo.svg" width="420">
  </picture>
</p>

<h1 align="center">Linkleaf CLI</h1>

<p align="center">
  Read/write <a href="https://developers.google.com/protocol-buffers">Protocol Buffers</a> feeds for <code>linkleaf.v1</code>.
</p>

---

## Overview

`linkleaf` is a tiny Go tool that round-trips a **Linkleaf Feed** between:

- **Binary protobuf** (`.pb`) – for compact, canonical publishing
- **JSON** – for easy editing, diffs, and tooling

**Schema:** [`proto/linkleaf/v1/feed.proto`](proto/linkleaf/v1/feed.proto)
**Go module:** `github.com/doriancodes/linkleaf-cli`
**Generated package import:** `github.com/doriancodes/linkleaf-cli/proto/linkleaf/v1`

---

## Quick start

```bash
# Clone and enter the repo
git clone https://github.com/doriancodes/linkleaf-cli.git
cd linkleaf-cli

# (1) Install the protobuf Go plugin (requires protoc installed on your system)
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
export PATH="$(go env GOPATH)/bin:$PATH"

# (2) Generate Go code from the proto (source-relative output)
protoc -I=proto --go_out=paths=source_relative:. proto/linkleaf/v1/feed.proto

# (3) Build the CLI
go build -o linkleaf ./cmd/linkleaf

# (4) Try it
./linkleaf read testdata/example.pb | jq .
```

## Usage

```bash
linkleaf – read/write Linkleaf Feed protobufs

Usage:
  linkleaf write <in.json> <out.pb>     JSON -> protobuf
  linkleaf read  <in.pb>                protobuf -> JSON (to stdout)

Notes:
  - JSON uses proto field names (snake_case).
  - "write" auto-fills:
      * feed.generated_at = now (UTC, RFC3339)
      * link.id if empty = sha256(url+"|"+date)[:12]
```
