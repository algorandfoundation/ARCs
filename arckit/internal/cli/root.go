package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	rootCmd.AddCommand(newSummaryCommand(opts, &exitCode, stdout))
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
				registry, registryDiagnostics, registryErr := adoption.LoadValidatedRegistry(root)
				diagnostics = append(diagnostics, registryDiagnostics...)
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
		Short: "Apply ARC and adoption formatting fixes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommand(opts, exitCode, stdout, func() (diag.Report, error) {
				files, fileErr := collectFmtTargets(args)
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
					if err := applyFixWithConfig(path, cfg); err != nil {
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
					Command:     "init arc",
					Created:     created,
					Diagnostics: diagnostics,
					Summary:     diag.Summarize(diagnostics),
					ExitCode:    diag.ExitCode(diagnostics),
				}
				if err != nil {
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

func newSummaryCommand(opts *options, exitCode *int, stdout io.Writer) *cobra.Command {
	command := &cobra.Command{
		Use:   "summary",
		Short: "Generate ARC editor repository summaries",
	}

	var outPath string
	repoCmd := &cobra.Command{
		Use:   "repo [repo-root]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Generate a repo-level ARC editor summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommand(opts, exitCode, stdout, func() (diag.Report, error) {
				if opts.Format == "json" {
					return newInvocationFailureReport("summary repo", errors.New("summary repo only supports text output in v1")), nil
				}

				root := "."
				if len(args) == 1 {
					root = args[0]
				}
				root = resolveRepoRoot(root)

				cfg, err := config.Load(root)
				if err != nil {
					return newConfigFailureReport("summary repo", root, err), nil
				}

				state, diagnostics, validateErr := repo.Validate(root, cfg)
				if validateErr != nil {
					return newInvocationFailureReport("summary repo", validateErr), nil
				}

				summary := repo.BuildSummary(state, diagnostics, time.Now())
				target := resolveSummaryOutputPath(root, outPath)
				if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
					return newInvocationFailureReport("summary repo", err), nil
				}
				if err := os.WriteFile(target, []byte(summary.Markdown()), 0o644); err != nil {
					return newInvocationFailureReport("summary repo", err), nil
				}

				return diag.Report{
					Command:  "summary repo",
					Created:  []string{target},
					ExitCode: 0,
				}, nil
			})
		},
	}
	repoCmd.Flags().StringVar(&outPath, "out", "", "output path for the generated summary markdown")
	command.AddCommand(repoCmd)

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
