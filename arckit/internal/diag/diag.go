package diag

import (
	"fmt"
	"sort"
)

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

type Origin string

const (
	OriginNative Origin = "native"
)

type Diagnostic struct {
	RuleID   string   `json:"rule_id"`
	Severity Severity `json:"severity"`
	Title    string   `json:"title"`
	Message  string   `json:"message"`
	Hint     string   `json:"hint"`
	Origin   Origin   `json:"origin"`
	File     string   `json:"file,omitempty"`
	Line     int      `json:"line,omitempty"`
	Column   int      `json:"column,omitempty"`
}

type Rule struct {
	ID          string   `json:"id"`
	Severity    Severity `json:"severity"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Rationale   string   `json:"rationale"`
	Hint        string   `json:"hint"`
	AutoFixable bool     `json:"auto_fixable"`
}

type Summary struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Info     int `json:"info"`
}

type Report struct {
	Command     string       `json:"command"`
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
	Summary     Summary      `json:"summary"`
	ExitCode    int          `json:"exit_code"`
	Rules       []Rule       `json:"rules,omitempty"`
	Rule        *Rule        `json:"rule,omitempty"`
	Created     []string     `json:"created,omitempty"`
}

func New(ruleID string, origin Origin, file string, line, column int, message string) Diagnostic {
	rule, ok := Lookup(ruleID)
	if !ok {
		return Diagnostic{
			RuleID:   ruleID,
			Severity: SeverityError,
			Title:    ruleID,
			Message:  message,
			Origin:   origin,
			File:     file,
			Line:     line,
			Column:   column,
		}
	}
	return Diagnostic{
		RuleID:   ruleID,
		Severity: rule.Severity,
		Title:    rule.Title,
		Message:  message,
		Hint:     rule.Hint,
		Origin:   origin,
		File:     file,
		Line:     line,
		Column:   column,
	}
}

func NewWithHint(ruleID string, origin Origin, file string, line, column int, message, hint string) Diagnostic {
	diagnostic := New(ruleID, origin, file, line, column, message)
	diagnostic.Hint = hint
	return diagnostic
}

func Summarize(diagnostics []Diagnostic) Summary {
	summary := Summary{}
	for _, diagnostic := range diagnostics {
		switch diagnostic.Severity {
		case SeverityError:
			summary.Errors++
		case SeverityWarning:
			summary.Warnings++
		case SeverityInfo:
			summary.Info++
		}
	}
	return summary
}

func ExitCode(diagnostics []Diagnostic) int {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == SeverityError {
			return 1
		}
	}
	return 0
}

func SortDiagnostics(diagnostics []Diagnostic) {
	sort.SliceStable(diagnostics, func(i, j int) bool {
		if diagnostics[i].File != diagnostics[j].File {
			return diagnostics[i].File < diagnostics[j].File
		}
		if diagnostics[i].Line != diagnostics[j].Line {
			return diagnostics[i].Line < diagnostics[j].Line
		}
		if diagnostics[i].Column != diagnostics[j].Column {
			return diagnostics[i].Column < diagnostics[j].Column
		}
		if diagnostics[i].RuleID != diagnostics[j].RuleID {
			return diagnostics[i].RuleID < diagnostics[j].RuleID
		}
		return diagnostics[i].Message < diagnostics[j].Message
	})
}

func FormatLocation(file string, line, column int) string {
	if file == "" {
		return ""
	}
	if line == 0 {
		return file
	}
	if column == 0 {
		return fmt.Sprintf("%s:%d", file, line)
	}
	return fmt.Sprintf("%s:%d:%d", file, line, column)
}
