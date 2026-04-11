package arc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func FindRepoRoot(start string) string {
	current := filepath.Clean(start)
	for {
		if current == string(filepath.Separator) || current == "" {
			break
		}
		if dirExists(filepath.Join(current, "ARCs")) || pathExists(filepath.Join(current, ".arckit.jsonc")) {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return filepath.Clean(start)
}

func withinRoot(root, path string) bool {
	_, err := relativeToRoot(root, path)
	return err == nil
}

func relativeToRoot(root, path string) (string, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(rootAbs, pathAbs)
	if err != nil {
		return "", err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path resolves outside root")
	}
	return relative, nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
