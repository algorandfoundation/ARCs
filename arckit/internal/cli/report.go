package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/algorandfoundation/ARCs/arckit/internal/config"
	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
)

func runCommand(opts *options, exitCode *int, stdout io.Writer, runner func() (diag.Report, error)) error {
	report, err := runner()
	if err != nil {
		return err
	}
	if report.Command == "" {
		report.Command = "arckit"
	}
	if report.Diagnostics != nil {
		diag.SortDiagnostics(report.Diagnostics)
		report.Summary = diag.Summarize(report.Diagnostics)
		if report.ExitCode == 0 {
			report.ExitCode = diag.ExitCode(report.Diagnostics)
		}
	}
	if err := render(stdout, report, *opts); err != nil {
		return err
	}
	*exitCode = report.ExitCode
	return nil
}

func reportForValidation(command string, diagnostics []diag.Diagnostic) diag.Report {
	report := diag.Report{
		Command:     command,
		Diagnostics: diagnostics,
	}
	report.Summary = diag.Summarize(diagnostics)
	report.ExitCode = diag.ExitCode(diagnostics)
	return report
}

func newInvocationFailureReport(command string, err error) diag.Report {
	diagnostic := diag.NewWithHint("R:026", diag.OriginNative, "", 0, 0, err.Error(), "Check the command help and provide the required arguments.")
	return diag.Report{
		Command:     command,
		Diagnostics: []diag.Diagnostic{diagnostic},
		Summary:     diag.Summarize([]diag.Diagnostic{diagnostic}),
		ExitCode:    2,
	}
}

func newConfigFailureReport(command, root string, err error) diag.Report {
	path := filepath.Join(filepath.Clean(root), config.FileName)
	diagnostic := diag.NewWithHint("R:028", diag.OriginNative, path, 0, 0, err.Error(), "Fix the .arckit.jsonc syntax, selectors, or rule IDs, then retry.")
	return diag.Report{
		Command:     command,
		Diagnostics: []diag.Diagnostic{diagnostic},
		Summary:     diag.Summarize([]diag.Diagnostic{diagnostic}),
		ExitCode:    2,
	}
}

func render(stdout io.Writer, report diag.Report, opts options) error {
	if opts.Format == "json" {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}
	writef := func(format string, args ...any) error {
		_, err := fmt.Fprintf(stdout, format, args...)
		return err
	}
	writeln := func(value string) error {
		_, err := fmt.Fprintln(stdout, value)
		return err
	}

	if report.Rule != nil {
		if err := writef("%s\n", report.Rule.ID); err != nil {
			return err
		}
		if err := writef("severity: %s\n", report.Rule.Severity); err != nil {
			return err
		}
		if err := writef("title: %s\n", report.Rule.Title); err != nil {
			return err
		}
		if err := writef("auto-fixable: %t\n\n", report.Rule.AutoFixable); err != nil {
			return err
		}
		if err := writef("%s\n\n", report.Rule.Description); err != nil {
			return err
		}
		if err := writef("Rationale: %s\n", report.Rule.Rationale); err != nil {
			return err
		}
		if err := writef("Hint: %s\n", report.Rule.Hint); err != nil {
			return err
		}
		return nil
	}
	if len(report.Rules) > 0 {
		for _, rule := range report.Rules {
			if err := writef("%s\t%s\tautofix=%t\t%s\n", rule.ID, rule.Severity, rule.AutoFixable, rule.Title); err != nil {
				return err
			}
		}
		return nil
	}
	if len(report.Created) > 0 {
		for _, path := range report.Created {
			if err := writeln(path); err != nil {
				return err
			}
		}
	}
	for _, diagnostic := range report.Diagnostics {
		location := diag.FormatLocation(diagnostic.File, diagnostic.Line, diagnostic.Column)
		if location != "" {
			if err := writef("%s %s %s: %s\n", strings.ToUpper(string(diagnostic.Severity)), diagnostic.RuleID, location, diagnostic.Message); err != nil {
				return err
			}
		} else {
			if err := writef("%s %s: %s\n", strings.ToUpper(string(diagnostic.Severity)), diagnostic.RuleID, diagnostic.Message); err != nil {
				return err
			}
		}
		if !opts.Quiet && diagnostic.Hint != "" {
			if err := writef("  hint: %s\n", diagnostic.Hint); err != nil {
				return err
			}
		}
	}
	if !opts.Quiet {
		if err := writef("summary: %d error(s), %d warning(s), %d info\n", report.Summary.Errors, report.Summary.Warnings, report.Summary.Info); err != nil {
			return err
		}
	}
	return nil
}
