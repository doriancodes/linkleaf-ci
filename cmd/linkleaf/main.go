package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "github.com/doriancodes/linkleaf-cli/proto/linkleaf/v1"
	"google.golang.org/protobuf/proto"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "init":
		cmdInit(os.Args[2:])
	case "add":
		cmdAdd(os.Args[2:])
	case "list":
		cmdList(os.Args[2:])
	case "print":
		cmdPrint(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `linkleaf – protobuf-only feed manager (linkleaf.v1)

Usage:
  linkleaf init  <file.pb> [-title "My Feed"] [-version 1]
  linkleaf add   -file <file.pb> -title "..." -url "..." -date YYYY-MM-DD [-summary "..."] [-tags a,b,c] [-via URL] [-id ID]
  linkleaf list  <file.pb>
  linkleaf print <file.pb>

Notes:
  • Data is stored ONLY in protobuf binary files (.pb).
  • "add" prepends links (newest first). If -id is empty: sha256(url+"|"+date)[:12].
  • "init" creates the file if it doesn't exist; "add" creates it on demand if needed.
`)
}

func cmdInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	var title string
	var version uint
	fs.StringVar(&title, "title", "", "feed title")
	fs.UintVar(&version, "version", 1, "feed version")
	fs.Parse(args)

	if fs.NArg() != 1 {
		usage()
		os.Exit(2)
	}
	path := fs.Arg(0)

	feed := &v1.Feed{
		Version:     uint32(version),
		Title:       title,
		GeneratedAt: nowRFC3339(),
		Links:       nil,
	}
	if err := saveFeed(path, feed); err != nil {
		die(err)
	}
	fmt.Printf("initialized %s (version=%d, title=%q)\n", path, version, title)
}

func cmdAdd(args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	var file, title, url, summary, tagsCSV, via, id, date string
	fs.StringVar(&file, "file", "", "protobuf feed file (.pb)")
	fs.StringVar(&title, "title", "", "link title (required)")
	fs.StringVar(&url, "url", "", "link URL (required)")
	fs.StringVar(&date, "date", "", "YYYY-MM-DD (required)")
	fs.StringVar(&summary, "summary", "", "short summary")
	fs.StringVar(&tagsCSV, "tags", "", "comma-separated tags (e.g. a,b,c)")
	fs.StringVar(&via, "via", "", "optional attribution URL")
	fs.StringVar(&id, "id", "", "stable ID (default: sha256(url|date)[:12])")
	fs.Parse(args)

	if file == "" || title == "" || url == "" || date == "" {
		fs.Usage()
		os.Exit(2)
	}

	feed, _ := loadFeed(file) // if not found, create a new feed
	if feed == nil {
		feed = &v1.Feed{}
	}
	feed.GeneratedAt = nowRFC3339()

	if id == "" {
		id = shortHash(url + "|" + date)
	}
	link := v1.Link{
		Id:      id,
		Title:   title,
		Url:     url,
		Summary: summary,
		Tags:    splitTags(tagsCSV),
		Date:    date,
		Via:     via,
	}

	// Prepend (newest first)
	feed.Links = append([]*v1.Link{&link}, feed.Links...)

	if err := saveFeed(file, feed); err != nil {
		die(err)
	}
	fmt.Printf("added [%s] %s\n", id, title)
}

func cmdList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	fs.Parse(args)
	if fs.NArg() != 1 {
		fs.Usage()
		os.Exit(2)
	}
	path := fs.Arg(0)

	feed, err := mustLoad(path)
	if err != nil {
		die(err)
	}
	fmt.Printf("Feed: %q  (version=%d, generated_at=%s)\n", feed.Title, feed.Version, feed.GeneratedAt)
	for i, l := range feed.Links {
		fmt.Printf("%3d) [%s] %s\n     %s\n     date=%s tags=%s\n",
			i+1, l.Id, l.Title, l.Url, l.Date, strings.Join(l.Tags, ","))
		if l.Summary != "" {
			fmt.Printf("     %s\n", wrap(l.Summary, 76, "     "))
		}
		if l.Via != "" {
			fmt.Printf("     via: %s\n", l.Via)
		}
	}
}

func cmdPrint(args []string) {
	// Human-friendly dump (no JSON); still loads from .pb
	fs := flag.NewFlagSet("print", flag.ExitOnError)
	fs.Parse(args)
	if fs.NArg() != 1 {
		fs.Usage()
		os.Exit(2)
	}
	path := fs.Arg(0)

	feed, err := mustLoad(path)
	if err != nil {
		die(err)
	}
	fmt.Printf("FEED\n----\nversion: %d\ntitle: %s\ngenerated_at: %s\nlinks: %d\n\n",
		feed.Version, feed.Title, feed.GeneratedAt, len(feed.Links))
	for _, l := range feed.Links {
		fmt.Printf("- id: %s\n  title: %s\n  url: %s\n  date: %s\n",
			l.Id, l.Title, l.Url, l.Date)
		if len(l.Tags) > 0 {
			fmt.Printf("  tags: %s\n", strings.Join(l.Tags, ", "))
		}
		if l.Summary != "" {
			fmt.Printf("  summary: %s\n", l.Summary)
		}
		if l.Via != "" {
			fmt.Printf("  via: %s\n", l.Via)
		}
		fmt.Println()
	}
}

// -------- storage (protobuf only) --------

func loadFeed(path string) (*v1.Feed, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	var feed v1.Feed
	if err := proto.Unmarshal(b, &feed); err != nil {
		return nil, fmt.Errorf("unmarshal protobuf: %w", err)
	}
	return &feed, nil
}

func mustLoad(path string) (*v1.Feed, error) {
	feed, err := loadFeed(path)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", path, err)
	}
	return feed, nil
}

func saveFeed(path string, feed *v1.Feed) error {
	b, err := proto.Marshal(feed)
	if err != nil {
		return fmt.Errorf("marshal protobuf: %w", err)
	}
	return writeFileAtomic(path, b, 0o644)
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
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

// -------- helpers --------

func nowRFC3339() string { return time.Now().UTC().Format(time.RFC3339) }

func shortHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:12]
}

func splitTags(csv string) []string {
	if strings.TrimSpace(csv) == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func wrap(s string, width int, indent string) string {
	if width <= 0 {
		return s
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}
	var b strings.Builder
	line := indent
	for _, w := range words {
		if len(line)+1+len(w) > len(indent)+width {
			b.WriteString(line)
			b.WriteByte('\n')
			line = indent + w
		} else {
			if len(line) > len(indent) {
				line += " "
			}
			line += w
		}
	}
	if line != "" {
		b.WriteString(line)
	}
	return b.String()
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
