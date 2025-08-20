package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	v1 "github.com/doriancodes/linkleaf-cli/proto/linkleaf/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func usage() {
	fmt.Fprintf(os.Stderr, `linkleaf – read/write Linkleaf Feed protobufs

Usage:
  linkleaf write <in.json> <out.pb>     JSON -> protobuf
  linkleaf read  <in.pb>                protobuf -> JSON (to stdout)

Notes:
  - JSON uses the proto field names (snake_case).
  - "write" auto-fills:
      * feed.generated_at = now (UTC, RFC3339)
      * link.id if empty = sha256(url+"|"+date)[:12]
`)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
		os.Exit(2)
	}

	switch flag.Arg(0) {
	case "write":
		if flag.NArg() != 3 {
			usage()
			os.Exit(2)
		}
		inJSON := flag.Arg(1)
		outPB := flag.Arg(2)
		if err := cmdWrite(inJSON, outPB); err != nil {
			die(err)
		}
	case "read":
		if flag.NArg() != 2 {
			usage()
			os.Exit(2)
		}
		inPB := flag.Arg(1)
		if err := cmdRead(inPB, os.Stdout); err != nil {
			die(err)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func cmdWrite(inJSON, outPB string) error {
	feed, err := readJSON(inJSON)
	if err != nil {
		return fmt.Errorf("read JSON: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if feed.GeneratedAt == "" {
		feed.GeneratedAt = now
	}
	// Fill missing link IDs deterministically from url|date.
	for i := range feed.Links {
		ln := feed.Links[i]
		if ln.Id == "" {
			base := ln.Url + "|" + ln.Date
			sum := sha256.Sum256([]byte(base))
			ln.Id = hex.EncodeToString(sum[:])[:12]
			feed.Links[i] = ln
		}
	}
	// Marshal protobuf binary and write atomically.
	b, err := proto.Marshal(feed)
	if err != nil {
		return fmt.Errorf("marshal protobuf: %w", err)
	}
	if err := writeFileAtomic(outPB, b, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", outPB, err)
	}
	return nil
}

func cmdRead(inPB string, out io.Writer) error {
	raw, err := os.ReadFile(inPB)
	if err != nil {
		return fmt.Errorf("read %s: %w", inPB, err)
	}
	var feed v1.Feed
	if err := proto.Unmarshal(raw, &feed); err != nil {
		return fmt.Errorf("unmarshal protobuf: %w", err)
	}
	// Stable JSON with proto (snake_case) field names and pretty output.
	m := protojson.MarshalOptions{
		UseProtoNames:   true,
		Indent:          "  ",
		EmitUnpopulated: false,
	}
	j, err := m.Marshal(&feed)
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	_, err = out.Write(append(j, '\n'))
	return err
}

func readJSON(path string) (*v1.Feed, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var feed v1.Feed
	u := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	if err := u.Unmarshal(raw, &feed); err != nil {
		return nil, fmt.Errorf("unmarshal JSON: %w", err)
	}
	return &feed, nil
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
