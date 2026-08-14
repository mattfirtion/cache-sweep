// cache-sweep finds and deletes node_modules directories and Python
// virtualenvs under a given directory tree, after listing them and
// confirming with the user.
package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type match struct {
	path  string
	bytes int64
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "cache-sweep:", err)
		os.Exit(1)
	}
}

func run() error {
	start := "."
	if len(os.Args) > 1 {
		start = os.Args[1]
	}

	root, err := filepath.Abs(start)
	if err != nil {
		return err
	}
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("%s: %w", root, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", root)
	}

	matches, err := discover(root)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		fmt.Println("No node_modules or Python virtualenvs found under", root)
		return nil
	}

	sort.Slice(matches, func(i, j int) bool { return matches[i].path < matches[j].path })

	var total int64
	fmt.Printf("Found %d directories under %s:\n\n", len(matches), root)
	for i, m := range matches {
		fmt.Printf("%3d. %-10s %s\n", i+1, humanSize(m.bytes), m.path)
		total += m.bytes
	}
	fmt.Printf("\nTotal reclaimable: %s\n\n", humanSize(total))

	if !confirm(fmt.Sprintf("Delete all %d directories (%s)? [y/N] ", len(matches), humanSize(total)), os.Stdin) {
		fmt.Println("Aborted, nothing deleted.")
		return nil
	}

	var reclaimed int64
	for _, m := range matches {
		if err := os.RemoveAll(m.path); err != nil {
			fmt.Fprintf(os.Stderr, "failed to remove %s: %v\n", m.path, err)
			continue
		}
		fmt.Println("removed", m.path)
		reclaimed += m.bytes
	}
	fmt.Printf("\nReclaimed %s.\n", humanSize(reclaimed))
	return nil
}

// discover walks root and returns node_modules directories and Python
// virtualenvs (directories containing pyvenv.cfg), pruning into any match
// so nested matches aren't double-counted.
func discover(root string) ([]match, error) {
	var matches []match

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Permission errors etc: skip and keep going.
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if path != root && d.Name() == "node_modules" {
			matches = append(matches, match{path: path, bytes: dirSize(path)})
			return fs.SkipDir
		}
		if isVenvDir(path) {
			matches = append(matches, match{path: path, bytes: dirSize(path)})
			return fs.SkipDir
		}
		return nil
	})
	return matches, err
}

func isVenvDir(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "pyvenv.cfg"))
	return err == nil
}

func dirSize(root string) int64 {
	var size int64
	filepath.WalkDir(root, func(_ string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if info, err := d.Info(); err == nil {
			size += info.Size()
		}
		return nil
	})
	return size
}

func humanSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func confirm(prompt string, in io.Reader) bool {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes"
}
