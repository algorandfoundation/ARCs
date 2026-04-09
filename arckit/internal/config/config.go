package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
)

const FileName = ".arckit.jsonc"

var (
	arcPathPattern = regexp.MustCompile(`(^|.*/)(ARCs|adoption|assets)/arc-(\d{4})(?:\.md|\.yaml|/.*)?$`)
	digitsPattern  = regexp.MustCompile(`^\d+$`)
)

type Config struct {
	ignoreArcs  map[int]struct{}
	ignoreRules map[string]struct{}
	ignoreByArc []arcRuleIgnore
}

type arcRuleIgnore struct {
	start int
	end   int
	rules map[string]struct{}
}

type rawConfig struct {
	IgnoreArcs  []any               `json:"ignoreArcs"`
	IgnoreRules []string            `json:"ignoreRules"`
	IgnoreByArc map[string][]string `json:"ignoreByArc"`
}

func Load(root string) (Config, error) {
	config := Config{
		ignoreArcs:  map[int]struct{}{},
		ignoreRules: map[string]struct{}{},
	}

	path := filepath.Join(filepath.Clean(root), FileName)
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return Config{}, fmt.Errorf("could not read %s: %w", path, err)
	}

	normalized, err := stripComments(content)
	if err != nil {
		return Config{}, fmt.Errorf("could not parse %s: %w", path, err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(normalized, &fields); err != nil {
		return Config{}, fmt.Errorf("could not parse %s: %w", path, err)
	}
	for key := range fields {
		switch key {
		case "ignoreArcs", "ignoreRules", "ignoreByArc":
		default:
			return Config{}, fmt.Errorf("unknown top-level key %q in %s", key, path)
		}
	}

	var raw rawConfig
	if err := json.Unmarshal(normalized, &raw); err != nil {
		return Config{}, fmt.Errorf("could not parse %s: %w", path, err)
	}

	for index, value := range raw.IgnoreArcs {
		number, err := parseARCValue(value)
		if err != nil {
			return Config{}, fmt.Errorf("invalid ignoreArcs[%d] in %s: %w", index, path, err)
		}
		config.ignoreArcs[number] = struct{}{}
	}

	for _, ruleID := range raw.IgnoreRules {
		if err := validateRuleID(ruleID); err != nil {
			return Config{}, fmt.Errorf("invalid ignoreRules entry in %s: %w", path, err)
		}
		config.ignoreRules[ruleID] = struct{}{}
	}

	for selector, rules := range raw.IgnoreByArc {
		start, end, err := parseSelector(selector)
		if err != nil {
			return Config{}, fmt.Errorf("invalid ignoreByArc selector %q in %s: %w", selector, path, err)
		}
		ruleSet := map[string]struct{}{}
		for _, ruleID := range rules {
			if err := validateRuleID(ruleID); err != nil {
				return Config{}, fmt.Errorf("invalid ignoreByArc value for %q in %s: %w", selector, path, err)
			}
			ruleSet[ruleID] = struct{}{}
		}
		config.ignoreByArc = append(config.ignoreByArc, arcRuleIgnore{
			start: start,
			end:   end,
			rules: ruleSet,
		})
	}

	return config, nil
}

func (config Config) IgnoreARC(number int) bool {
	if number < 0 {
		return false
	}
	_, ok := config.ignoreArcs[number]
	return ok
}

func (config Config) IgnoreRule(ruleID string) bool {
	_, ok := config.ignoreRules[ruleID]
	return ok
}

func (config Config) IgnoreRuleForARC(ruleID string, number int) bool {
	if number < 0 {
		return false
	}
	for _, entry := range config.ignoreByArc {
		if number < entry.start || number > entry.end {
			continue
		}
		if _, ok := entry.rules[ruleID]; ok {
			return true
		}
	}
	return false
}

func (config Config) FilterDiagnostics(diagnostics []diag.Diagnostic) []diag.Diagnostic {
	if len(diagnostics) == 0 {
		return diagnostics
	}
	filtered := make([]diag.Diagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		if config.IgnoreRule(diagnostic.RuleID) {
			continue
		}
		number, ok := ARCNumberForPath(diagnostic.File)
		if ok && config.IgnoreARC(number) {
			continue
		}
		if ok && config.IgnoreRuleForARC(diagnostic.RuleID, number) {
			continue
		}
		filtered = append(filtered, diagnostic)
	}
	return filtered
}

func ARCNumberForPath(path string) (int, bool) {
	matches := arcPathPattern.FindStringSubmatch(filepath.ToSlash(filepath.Clean(path)))
	if len(matches) != 4 {
		return 0, false
	}
	number, err := strconv.Atoi(matches[3])
	if err != nil {
		return 0, false
	}
	return number, true
}

func stripComments(input []byte) ([]byte, error) {
	output := make([]byte, 0, len(input))
	inString := false
	lineComment := false
	blockComment := false
	escaped := false

	for index := 0; index < len(input); index++ {
		current := input[index]

		if lineComment {
			if current == '\n' {
				lineComment = false
				output = append(output, current)
			}
			continue
		}

		if blockComment {
			if current == '*' && index+1 < len(input) && input[index+1] == '/' {
				blockComment = false
				index++
				continue
			}
			if current == '\n' || current == '\r' || current == '\t' {
				output = append(output, current)
			} else {
				output = append(output, ' ')
			}
			continue
		}

		if inString {
			output = append(output, current)
			if escaped {
				escaped = false
				continue
			}
			if current == '\\' {
				escaped = true
				continue
			}
			if current == '"' {
				inString = false
			}
			continue
		}

		if current == '"' {
			inString = true
			output = append(output, current)
			continue
		}

		if current == '/' && index+1 < len(input) {
			switch input[index+1] {
			case '/':
				lineComment = true
				index++
				continue
			case '*':
				blockComment = true
				index++
				continue
			}
		}

		output = append(output, current)
	}

	if blockComment {
		return nil, errors.New("unterminated block comment")
	}

	return output, nil
}

func parseARCValue(value any) (int, error) {
	switch typed := value.(type) {
	case float64:
		if typed != math.Trunc(typed) {
			return 0, fmt.Errorf("expected an integer ARC number, got %v", typed)
		}
		if typed < 0 {
			return 0, fmt.Errorf("ARC numbers must be non-negative, got %v", typed)
		}
		return int(typed), nil
	case string:
		return parseARCString(typed)
	default:
		return 0, fmt.Errorf("expected an integer ARC number, got %T", value)
	}
}

func parseSelector(selector string) (int, int, error) {
	parts := strings.Split(strings.TrimSpace(selector), "-")
	switch len(parts) {
	case 1:
		number, err := parseARCString(parts[0])
		if err != nil {
			return 0, 0, err
		}
		return number, number, nil
	case 2:
		start, err := parseARCString(parts[0])
		if err != nil {
			return 0, 0, err
		}
		end, err := parseARCString(parts[1])
		if err != nil {
			return 0, 0, err
		}
		if start > end {
			return 0, 0, fmt.Errorf("range start %d is greater than end %d", start, end)
		}
		return start, end, nil
	default:
		return 0, 0, fmt.Errorf("expected ARC selector N or N-M")
	}
}

func parseARCString(value string) (int, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("expected a non-empty ARC selector")
	}
	if !digitsPattern.MatchString(trimmed) {
		return 0, fmt.Errorf("expected digits only, got %q", value)
	}
	number, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, err
	}
	if number < 0 {
		return 0, fmt.Errorf("ARC numbers must be non-negative, got %d", number)
	}
	return number, nil
}

func validateRuleID(ruleID string) error {
	if _, ok := diag.Lookup(ruleID); !ok {
		return fmt.Errorf("unknown rule %q", ruleID)
	}
	return nil
}
