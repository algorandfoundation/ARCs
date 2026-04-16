package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func CopyDir(t testing.TB, src string) string {
	t.Helper()

	dst := t.TempDir()
	if err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, relative)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, content, info.Mode())
	}); err != nil {
		t.Fatalf("CopyDir() error = %v", err)
	}

	return dst
}

func WriteFile(t testing.TB, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}

func WriteTrimmedFile(t testing.TB, path string, content string) {
	t.Helper()
	WriteFile(t, path, strings.TrimSpace(content)+"\n")
}
