package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/algorandfoundation/ARCs/arckit/internal/adoption"
	"github.com/algorandfoundation/ARCs/arckit/internal/arc"
	"github.com/algorandfoundation/ARCs/arckit/internal/config"
	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
	"github.com/algorandfoundation/ARCs/arckit/internal/repo"
	"github.com/algorandfoundation/ARCs/arckit/internal/scaffold"
	"github.com/algorandfoundation/ARCs/arckit/internal/transition"
	"github.com/spf13/cobra"
)

type options struct {
	Format string
	Quiet  bool
}

func Execute() int {
	return ExecuteArgs(os.Args[1:], os.Stdout, os.Stderr)
}

func ExecuteArgs(args []string, stdout io.Writer, stderr io.Writer) int {
	opts := &options{
		Format: "text",
	}
	exitCode := 0

	rootCmd := &cobra.Command{
		Use:           "arckit",
		Short:         "Validate and scaffold ARC repository artifacts",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs(args)
	rootCmd.PersistentFlags().StringVar(&opts.Format, "format", "text", "output format: text or json")
	rootCmd.PersistentFlags().BoolVar(&opts.Quiet, "quiet", false, "suppress non-essential text output")

	rootCmd.AddCommand(newValidateCommand(opts, &exitCode, stdout))
	rootCmd.AddCommand(newFmtCommand(opts, &exitCode, stdout))
	rootCmd.AddCommand(newInitCommand(opts, &exitCode, stdout))
	rootCmd.AddCommand(newRulesCommand(opts, &exitCode, stdout))
	rootCmd.AddCommand(newExplainCommand(opts, &exitCode, stdout))

	if err := rootCmd.Execute(); err != nil {
		report := newInvocationFailureReport("arckit", err)
		if renderErr := render(stderr, report, *opts); renderErr != nil {
			return 2
		}
		return 2
	}
	return exitCode
}

func newValidateCommand(opts *options, exitCode *int, stdout io.Writer) *cobra.Command {
	var ignoreConfig bool
	var enforceRules []string
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate ARC repository artifacts",
	}
	validateCmd.PersistentFlags().BoolVar(&ignoreConfig, "ignore-config", false, "do not load .arckit.jsonc suppressions")
	validateCmd.PersistentFlags().StringSliceVar(&enforceRules, "enforce-rule", nil, "validate these rule IDs even if .arckit.jsonc suppresses them")

	validateCmd.AddCommand(&cobra.Command{
		Use:   "arc <arc-file>",
		Args:  cobra.ExactArgs(1),
		Short: "Validate one ARC file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommand(opts, exitCode, stdout, func() (diag.Report, error) {
				root := arc.FindRepoRoot(filepath.Dir(args[0]))
				cfg, configErr := loadValidationConfig(root, ignoreConfig, enforceRules)
				if configErr != nil {
					return newConfigFailureReport("validate arc", root, configErr), nil
				}
				if shouldIgnorePath(cfg, args[0]) {
					return reportForValidation("validate arc", nil), nil
				}

				document, diagnostics, loadErr := arc.Load(args[0])
				if loadErr != nil {
					return reportForValidation("validate arc", cfg.FilterDiagnostics(diagnostics)), nil
				}
				diagnostics = append(diagnostics, arc.Validate(document, root)...)
				return reportForValidation("validate arc", cfg.FilterDiagnostics(diagnostics)), nil
			})
		},
	})

	validateCmd.AddCommand(&cobra.Command{
		Use:   "adoption <adoption-file>",
		Args:  cobra.ExactArgs(1),
		Short: "Validate one adoption summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommand(opts, exitCode, stdout, func() (diag.Report, error) {
				root := arc.FindRepoRoot(filepath.Dir(args[0]))
				cfg, configErr := loadValidationConfig(root, ignoreConfig, enforceRules)
				if configErr != nil {
					return newConfigFailureReport("validate adoption", root, configErr), nil
				}
				if shouldIgnorePath(cfg, args[0]) {
					return reportForValidation("validate adoption", nil), nil
				}

				summary, diagnostics, loadErr := adoption.Load(args[0])
				if loadErr != nil {
					return reportForValidation("validate adoption", cfg.FilterDiagnostics(diagnostics)), nil
				}
				registry, registryDiagnostics, registryErr := adoption.LoadVettedAdopters(adoption.RegistryPath(root))
				diagnostics = append(diagnostics, registryDiagnostics...)
				if registryErr == nil && len(registryDiagnostics) == 0 {
					registryDiagnostics = adoption.ValidateVettedAdopters(registry)
					diagnostics = append(diagnostics, registryDiagnostics...)
				}
				if registryErr != nil || len(registryDiagnostics) != 0 {
					registry = nil
				}
				var document *arc.Document
				if summary.Arc != 0 {
					arcPath := filepath.Join(root, "ARCs", fmt.Sprintf("arc-%04d.md", summary.Arc))
					if loaded, loadDiagnostics, err := arc.Load(arcPath); err == nil {
						document = loaded
						diagnostics = append(diagnostics, loadDiagnostics...)
						diagnostics = append(diagnostics, arc.Validate(document, root)...)
					} else {
						diagnostics = append(diagnostics, loadDiagnostics...)
						if errors.Is(err, os.ErrNotExist) {
							diagnostics = append(diagnostics, diag.NewWithHint("R:018", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("orphaned adoption summary for ARC %d", summary.Arc), "Remove the orphaned adoption summary or add the matching ARC file."))
						}
					}
				}
				diagnostics = append(diagnostics, adoption.Validate(summary, document, registry)...)
				return reportForValidation("validate adoption", cfg.FilterDiagnostics(diagnostics)), nil
			})
		},
	})

	validateCmd.AddCommand(&cobra.Command{
		Use:   "links <path...>",
		Args:  cobra.MinimumNArgs(1),
		Short: "Validate ARC-local links in ARC files",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommand(opts, exitCode, stdout, func() (diag.Report, error) {
				root := resolveRepoRoot(args[0])
				cfg, configErr := loadValidationConfig(root, ignoreConfig, enforceRules)
				if configErr != nil {
					return newConfigFailureReport("validate links", root, configErr), nil
				}
				files, fileErr := collectARCFiles(args)
				if fileErr != nil {
					return newInvocationFailureReport("validate links", fileErr), nil
				}
				diagnostics := make([]diag.Diagnostic, 0)
				for _, path := range files {
					if shouldIgnorePath(cfg, path) {
						continue
					}
					document, _, err := arc.Load(path)
					if err != nil {
						diagnostics = append(diagnostics, diag.NewWithHint("R:027", diag.OriginNative, path, 0, 0, err.Error(), "Check the file path and permissions, then retry."))
						continue
					}
					diagnostics = append(diagnostics, arc.ValidateLinks(document, arc.FindRepoRoot(filepath.Dir(path)))...)
				}
				return reportForValidation("validate links", cfg.FilterDiagnostics(diagnostics)), nil
			})
		},
	})

	validateCmd.AddCommand(&cobra.Command{
		Use:   "repo [repo-root]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Validate the repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommand(opts, exitCode, stdout, func() (diag.Report, error) {
				root := "."
				if len(args) == 1 {
					root = args[0]
				}
				root = resolveRepoRoot(root)
				cfg, configErr := loadValidationConfig(root, ignoreConfig, enforceRules)
				if configErr != nil {
					return newConfigFailureReport("validate repo", root, configErr), nil
				}
				_, diagnostics, _ := repo.Validate(root, cfg)
				return reportForValidation("validate repo", diagnostics), nil
			})
		},
	})

	transitionCmd := &cobra.Command{
		Use:   "transition <arc-file>",
		Args:  cobra.ExactArgs(1),
		Short: "Validate a machine-verifiable status transition",
	}
	var targetStatus string
	transitionCmd.Flags().StringVar(&targetStatus, "to", "", "target ARC status")
	if err := transitionCmd.MarkFlagRequired("to"); err != nil {
		panic(err)
	}
	transitionCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runCommand(opts, exitCode, stdout, func() (diag.Report, error) {
			if strings.TrimSpace(targetStatus) == "" {
				return newInvocationFailureReport("validate transition", errors.New("missing --to status")), nil
			}
			root := arc.FindRepoRoot(filepath.Dir(args[0]))
			cfg, configErr := loadValidationConfig(root, ignoreConfig, enforceRules)
			if configErr != nil {
				return newConfigFailureReport("validate transition", root, configErr), nil
			}
			if shouldIgnorePath(cfg, args[0]) {
				return reportForValidation("validate transition", nil), nil
			}
			diagnostics, _ := transition.Validate(args[0], targetStatus)
			return reportForValidation("validate transition", cfg.FilterDiagnostics(diagnostics)), nil
		})
	}
	validateCmd.AddCommand(transitionCmd)
	return validateCmd
}

func loadValidationConfig(root string, ignoreConfig bool, enforceRules []string) (config.Config, error) {
	if ignoreConfig {
		return config.Config{}, nil
	}
	cfg, err := config.Load(root)
	if err != nil {
		return config.Config{}, err
	}
	for _, ruleID := range enforceRules {
		if _, ok := diag.Lookup(ruleID); !ok {
			return config.Config{}, fmt.Errorf("invalid enforced rule %q: unknown rule", ruleID)
		}
		cfg = cfg.WithRuleEnforced(ruleID)
	}
	return cfg, nil
}

func newFmtCommand(opts *options, exitCode *int, stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "fmt <path...>",
		Args:  cobra.MinimumNArgs(1),
		Short: "Apply ARC front matter formatting fixes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommand(opts, exitCode, stdout, func() (diag.Report, error) {
				files, fileErr := collectARCFiles(args)
				if fileErr != nil {
					return newInvocationFailureReport("fmt", fileErr), nil
				}
				diagnostics := make([]diag.Diagnostic, 0)
				configs := map[string]config.Config{}
				for _, path := range files {
					root := arc.FindRepoRoot(filepath.Dir(path))
					cfg, ok := configs[root]
					if !ok {
						loaded, configErr := config.Load(root)
						if configErr != nil {
							diagnostics = append(diagnostics, diag.NewWithHint("R:027", diag.OriginNative, path, 0, 0, configErr.Error(), "Check the file permissions and content, then retry."))
							continue
						}
						cfg = loaded
						configs[root] = cfg
					}
					if shouldIgnorePath(cfg, path) {
						continue
					}
					if err := applyNativeFixWithConfig(path, cfg); err != nil {
						diagnostics = append(diagnostics, diag.NewWithHint("R:027", diag.OriginNative, path, 0, 0, err.Error(), "Check the file permissions and content, then retry."))
					}
				}
				report := reportForValidation("fmt", diagnostics)
				return report, nil
			})
		},
	}
}

func newInitCommand(opts *options, exitCode *int, stdout io.Writer) *cobra.Command {
	var initOptions scaffold.InitOptions
	command := &cobra.Command{
		Use:   "init",
		Short: "Scaffold ARC artifacts",
	}
	initArcCmd := &cobra.Command{
		Use:   "arc",
		Short: "Scaffold a new ARC",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommand(opts, exitCode, stdout, func() (diag.Report, error) {
				if initOptions.Number <= 0 || strings.TrimSpace(initOptions.Title) == "" || strings.TrimSpace(initOptions.Type) == "" || strings.TrimSpace(initOptions.Sponsor) == "" {
					return newInvocationFailureReport("init arc", errors.New("number, title, type, and sponsor are required")), nil
				}
				if initOptions.Root == "" {
					initOptions.Root = "."
				}
				created, diagnostics, err := scaffold.InitARC(initOptions)
				report := diag.Report{
					Command:  "init arc",
					Created:  created,
					Summary:  diag.Summarize(diagnostics),
					ExitCode: diag.ExitCode(diagnostics),
				}
				if err != nil {
					report.Diagnostics = diagnostics
					report.Summary = diag.Summarize(report.Diagnostics)
					report.ExitCode = 2
				}
				return report, nil
			})
		},
	}
	initArcCmd.Flags().StringVar(&initOptions.Root, "root", ".", "repository root")
	initArcCmd.Flags().IntVar(&initOptions.Number, "number", 0, "ARC number")
	initArcCmd.Flags().StringVar(&initOptions.Title, "title", "", "ARC title")
	initArcCmd.Flags().StringVar(&initOptions.Type, "type", "", "ARC type")
	initArcCmd.Flags().StringVar(&initOptions.Category, "category", "", "ARC category")
	initArcCmd.Flags().StringVar(&initOptions.SubCategory, "sub-category", "", "ARC sub-category")
	initArcCmd.Flags().StringVar(&initOptions.Sponsor, "sponsor", "", "ARC sponsor")
	initArcCmd.Flags().StringVar(&initOptions.Author, "author", "", "ARC author")
	initArcCmd.Flags().StringVar(&initOptions.Description, "description", "", "ARC description")
	initArcCmd.Flags().BoolVar(&initOptions.ImplementationRequired, "implementation-required", false, "whether the ARC requires a reference implementation")
	command.AddCommand(initArcCmd)
	return command
}

func newRulesCommand(opts *options, exitCode *int, stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "rules",
		Short: "List known rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommand(opts, exitCode, stdout, func() (diag.Report, error) {
				report := diag.Report{
					Command:  "rules",
					Rules:    diag.AllRules(),
					ExitCode: 0,
				}
				return report, nil
			})
		},
	}
}

func newExplainCommand(opts *options, exitCode *int, stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "explain <rule-id>",
		Args:  cobra.ExactArgs(1),
		Short: "Explain one rule",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommand(opts, exitCode, stdout, func() (diag.Report, error) {
				rule, ok := diag.Lookup(args[0])
				if !ok {
					return newInvocationFailureReport("explain", fmt.Errorf("unknown rule %q", args[0])), nil
				}
				return diag.Report{
					Command:  "explain",
					Rule:     &rule,
					ExitCode: 0,
				}, nil
			})
		},
	}
}

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

func collectARCFiles(paths []string) ([]string, error) {
	pattern := regexp.MustCompile(`(^|.*/)ARCs/arc-\d{4}\.md$`)
	seen := map[string]struct{}{}
	arcFiles := make([]string, 0)
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
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
					return nil
				}
				if !pattern.MatchString(filepath.ToSlash(walkPath)) {
					return nil
				}
				cleaned := filepath.Clean(walkPath)
				if _, ok := seen[cleaned]; ok {
					return nil
				}
				seen[cleaned] = struct{}{}
				arcFiles = append(arcFiles, cleaned)
				return nil
			})
			if err != nil {
				return nil, err
			}
			continue
		}
		if !pattern.MatchString(filepath.ToSlash(path)) {
			return nil, fmt.Errorf("%s is not an ARC Markdown file under ARCs/arc-####.md", path)
		}
		cleaned := filepath.Clean(path)
		if _, ok := seen[cleaned]; ok {
			continue
		}
		seen[cleaned] = struct{}{}
		arcFiles = append(arcFiles, cleaned)
	}
	if len(arcFiles) == 0 {
		return nil, errors.New("no ARC Markdown files found under ARCs/arc-####.md")
	}
	sort.Strings(arcFiles)
	return arcFiles, nil
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
