package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/algorandfoundation/ARCs/arckit/internal/arc"
	"github.com/algorandfoundation/ARCs/arckit/internal/config"
)

var (
	arcFilePattern         = regexp.MustCompile(`(^|.*/)ARCs/arc-\d{4}\.md$`)
	adoptionSummaryPattern = regexp.MustCompile(`(^|.*/)adoption/arc-\d{4}\.yaml$`)
)

func isAdoptionSummaryPath(path string) bool {
	slashPath := filepath.ToSlash(filepath.Clean(path))
	slashPath = strings.ReplaceAll(slashPath, `\`, "/")
	return adoptionSummaryPattern.MatchString(slashPath)
}

func collectFmtTargets(paths []string) ([]string, error) {
	return collectMatchingPaths(
		paths,
		func(slashPath string, entry os.DirEntry) bool {
			return !entry.IsDir() && ((strings.HasSuffix(entry.Name(), ".md") && arcFilePattern.MatchString(slashPath)) || isAdoptionSummaryPath(slashPath))
		},
		func(path string) error {
			if arcFilePattern.MatchString(path) || isAdoptionSummaryPath(path) {
				return nil
			}
			return fmt.Errorf("%s is not an ARC Markdown file under ARCs/arc-####.md or an adoption summary under adoption/arc-####.yaml", path)
		},
		"no ARC Markdown files or adoption summaries found under ARCs/arc-####.md or adoption/arc-####.yaml",
	)
}

func collectARCFiles(paths []string) ([]string, error) {
	return collectMatchingPaths(
		paths,
		func(slashPath string, entry os.DirEntry) bool {
			return !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") && arcFilePattern.MatchString(slashPath)
		},
		func(path string) error {
			if arcFilePattern.MatchString(path) {
				return nil
			}
			return fmt.Errorf("%s is not an ARC Markdown file under ARCs/arc-####.md", path)
		},
		"no ARC Markdown files found under ARCs/arc-####.md",
	)
}

func collectMatchingPaths(
	paths []string,
	include func(slashPath string, entry os.DirEntry) bool,
	validateFile func(slashPath string) error,
	emptyMessage string,
) ([]string, error) {
	seen := map[string]struct{}{}
	out := make([]string, 0)

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			err := filepath.WalkDir(path, func(walkPath string, entry os.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if !include(filepath.ToSlash(walkPath), entry) {
					return nil
				}
				appendUniquePath(&out, seen, walkPath)
				return nil
			})
			if err != nil {
				return nil, err
			}
			continue
		}

		slashPath := filepath.ToSlash(path)
		if err := validateFile(slashPath); err != nil {
			return nil, err
		}
		appendUniquePath(&out, seen, path)
	}

	if len(out) == 0 {
		return nil, errors.New(emptyMessage)
	}
	sort.Strings(out)
	return out, nil
}

func appendUniquePath(out *[]string, seen map[string]struct{}, path string) {
	cleaned := filepath.Clean(path)
	if _, ok := seen[cleaned]; ok {
		return
	}
	seen[cleaned] = struct{}{}
	*out = append(*out, cleaned)
}

func shouldIgnorePath(cfg config.Config, path string) bool {
	number, ok := config.ARCNumberForPath(path)
	return ok && cfg.IgnoreARC(number)
}

func resolveRepoRoot(path string) string {
	cleaned := filepath.Clean(path)
	if dirExists(filepath.Join(cleaned, "ARCs")) && dirExists(filepath.Join(cleaned, "adoption")) {
		return cleaned
	}
	return arc.FindRepoRoot(cleaned)
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func resolveSummaryOutputPath(root string, outPath string) string {
	if strings.TrimSpace(outPath) == "" {
		return filepath.Join(filepath.Clean(root), "arc-summary.md")
	}
	if filepath.IsAbs(outPath) {
		return filepath.Clean(outPath)
	}
	return filepath.Join(filepath.Clean(root), filepath.Clean(outPath))
}
