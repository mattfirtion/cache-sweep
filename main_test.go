package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestHumanSize(t *testing.T) {
	cases := []struct {
		bytes int64
		want  string
	}{
		{0, "0B"},
		{27, "27B"},
		{1023, "1023B"},
		{1024, "1.0KiB"},
		{1536, "1.5KiB"},
		{1024 * 1024, "1.0MiB"},
		{1024 * 1024 * 1024, "1.0GiB"},
	}
	for _, c := range cases {
		if got := humanSize(c.bytes); got != c.want {
			t.Errorf("humanSize(%d) = %q, want %q", c.bytes, got, c.want)
		}
	}
}

func TestIsVenvDir(t *testing.T) {
	dir := t.TempDir()
	if isVenvDir(dir) {
		t.Fatalf("isVenvDir(%s) = true, want false before pyvenv.cfg exists", dir)
	}
	if err := os.WriteFile(filepath.Join(dir, "pyvenv.cfg"), []byte("home = /usr/bin"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !isVenvDir(dir) {
		t.Fatalf("isVenvDir(%s) = false, want true after pyvenv.cfg exists", dir)
	}
}

func TestDirSize(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.txt"), "12345")
	writeFile(t, filepath.Join(dir, "nested", "b.txt"), "1234567890")

	got := dirSize(dir)
	want := int64(5 + 10)
	if got != want {
		t.Errorf("dirSize(%s) = %d, want %d", dir, got, want)
	}
}

func TestDiscover(t *testing.T) {
	root := t.TempDir()

	// node_modules with a nested node_modules that must be pruned, not
	// double-counted.
	writeFile(t, filepath.Join(root, "proj", "node_modules", "pkg", "index.js"), "console.log(1)")
	writeFile(t, filepath.Join(root, "proj", "node_modules", "pkg", "nested", "node_modules", "f.js"), "x")

	// Python virtualenv, identified by pyvenv.cfg regardless of dir name.
	writeFile(t, filepath.Join(root, "proj", ".venv", "pyvenv.cfg"), "home = /usr/bin")
	writeFile(t, filepath.Join(root, "proj", ".venv", "lib", "site.py"), "y")

	// Untouched source dir, should not be reported.
	writeFile(t, filepath.Join(root, "proj", "src", "main.go"), "package main")

	matches, err := discover(root)
	if err != nil {
		t.Fatal(err)
	}

	var paths []string
	for _, m := range matches {
		paths = append(paths, m.path)
	}
	sort.Strings(paths)

	want := []string{
		filepath.Join(root, "proj", ".venv"),
		filepath.Join(root, "proj", "node_modules"),
	}
	if len(paths) != len(want) {
		t.Fatalf("discover() found %v, want %v", paths, want)
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Errorf("discover()[%d] = %s, want %s", i, paths[i], want[i])
		}
	}

	for _, m := range matches {
		if m.path == filepath.Join(root, "proj", "node_modules") && m.bytes != int64(len("console.log(1)")+len("x")) {
			t.Errorf("node_modules size = %d, want %d (nested node_modules should be included in its size but not reported separately)", m.bytes, len("console.log(1)")+len("x"))
		}
	}
}

func TestConfirm(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"y\n", true},
		{"Y\n", true},
		{"yes\n", true},
		{"YES\n", true},
		{"n\n", false},
		{"no\n", false},
		{"\n", false},
		{"", false},
		{"maybe\n", false},
	}
	for _, c := range cases {
		if got := confirm("prompt", strings.NewReader(c.input)); got != c.want {
			t.Errorf("confirm(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
