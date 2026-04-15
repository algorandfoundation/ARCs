package scaffold

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/algorandfoundation/ARCs/arckit/internal/adoption"
	"github.com/algorandfoundation/ARCs/arckit/internal/arc"
	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
)

type InitOptions struct {
	Root                   string
	Number                 int
	Title                  string
	Type                   string
	Category               string
	SubCategory            string
	Sponsor                string
	Author                 string
	Description            string
	ImplementationRequired bool
}

func InitARC(options InitOptions) ([]string, []diag.Diagnostic, error) {
	root := filepath.Clean(options.Root)
	number := fmt.Sprintf("%04d", options.Number)
	arcPath := filepath.Join(root, "ARCs", "arc-"+number+".md")
	adoptionPath := filepath.Join(root, "adoption", "arc-"+number+".yaml")
	assetPath := filepath.Join(root, "assets", "arc-"+number)
	registryPath := filepath.Join(root, "adoption", adoption.VettedAdoptersFileName)
	created := []string{arcPath, adoptionPath, assetPath}

	if diagnostics := arc.ValidateCategoryMetadata(arcPath, 1, 1, strings.TrimSpace(options.Category), strings.TrimSpace(options.SubCategory)); len(diagnostics) != 0 {
		return nil, diagnostics, nil
	}

	for _, path := range []string{arcPath, adoptionPath, assetPath} {
		if _, err := os.Stat(path); err == nil {
			return nil, []diag.Diagnostic{
				diag.NewWithHint("R:027", diag.OriginNative, path, 0, 0, "refusing to overwrite existing scaffold target", "Choose a new ARC number or remove the existing files first."),
			}, fmt.Errorf("%s already exists", path)
		}
	}

	if err := os.MkdirAll(filepath.Dir(arcPath), 0o755); err != nil {
		return nil, []diag.Diagnostic{diag.NewWithHint("R:027", diag.OriginNative, arcPath, 0, 0, err.Error(), "Create the ARCs directory and retry.")}, err
	}
	if err := os.MkdirAll(filepath.Dir(adoptionPath), 0o755); err != nil {
		return nil, []diag.Diagnostic{diag.NewWithHint("R:027", diag.OriginNative, adoptionPath, 0, 0, err.Error(), "Create the adoption directory and retry.")}, err
	}
	if err := os.MkdirAll(assetPath, 0o755); err != nil {
		return nil, []diag.Diagnostic{diag.NewWithHint("R:027", diag.OriginNative, assetPath, 0, 0, err.Error(), "Create the asset directory and retry.")}, err
	}

	author := strings.TrimSpace(options.Author)
	if author == "" {
		author = "TBD (@todo)"
	}
	description := strings.TrimSpace(options.Description)
	if description == "" {
		description = "TODO: add a short description."
	}
	now := time.Now().UTC().Format("2006-01-02")

	arcContent := renderARC(options, author, description, now)
	adoptionContent := renderAdoption(options, now)

	if err := os.WriteFile(arcPath, []byte(arcContent), 0o644); err != nil {
		return nil, []diag.Diagnostic{diag.NewWithHint("R:027", diag.OriginNative, arcPath, 0, 0, err.Error(), "Check filesystem permissions and retry.")}, err
	}
	if err := os.WriteFile(adoptionPath, []byte(adoptionContent), 0o644); err != nil {
		return nil, []diag.Diagnostic{diag.NewWithHint("R:027", diag.OriginNative, adoptionPath, 0, 0, err.Error(), "Check filesystem permissions and retry.")}, err
	}
	if _, err := os.Stat(registryPath); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(registryPath, []byte(renderVettedAdopters()), 0o644); err != nil {
			return nil, []diag.Diagnostic{diag.NewWithHint("R:027", diag.OriginNative, registryPath, 0, 0, err.Error(), "Check filesystem permissions and retry.")}, err
		}
		created = append(created, registryPath)
	} else if err != nil {
		return nil, []diag.Diagnostic{diag.NewWithHint("R:027", diag.OriginNative, registryPath, 0, 0, err.Error(), "Check filesystem permissions and retry.")}, err
	}
	return created, nil, nil
}

func renderARC(options InitOptions, author string, description string, now string) string {
	number := fmt.Sprintf("%04d", options.Number)
	lines := []string{
		"---",
		fmt.Sprintf("arc: %d", options.Number),
		fmt.Sprintf("title: %s", options.Title),
		fmt.Sprintf("description: %s", description),
		"author:",
		fmt.Sprintf("  - %s", author),
		"discussions-to: https://github.com/algorandfoundation/ARCs/discussions/0",
		"status: Draft",
		fmt.Sprintf("type: %s", options.Type),
	}
	if category := strings.TrimSpace(options.Category); category != "" {
		lines = append(lines, fmt.Sprintf("category: %s", category))
	}
	if subCategory := strings.TrimSpace(options.SubCategory); subCategory != "" {
		lines = append(lines, fmt.Sprintf("sub-category: %s", subCategory))
	}
	lines = append(lines,
		fmt.Sprintf("created: %s", now),
		fmt.Sprintf("sponsor: %s", options.Sponsor),
		fmt.Sprintf("implementation-required: %t", options.ImplementationRequired),
		fmt.Sprintf("adoption-summary: adoption/arc-%s.yaml", number),
		"---",
		"",
		"## Abstract",
		"",
		"TODO: add the abstract.",
		"",
		"## Motivation",
		"",
		"TODO: explain the problem or opportunity.",
		"",
		"## Specification",
		"",
		"TODO: define the proposed behavior.",
		"",
		"## Rationale",
		"",
		"TODO: explain the design tradeoffs.",
		"",
		"## Security Considerations",
		"",
		"TODO: document the relevant security considerations.",
		"",
		"## Copyright",
		"",
		"Copyright and related rights waived via CC0 1.0.",
		"",
	)
	return strings.Join(lines, "\n")
}

func renderAdoption(options InitOptions, now string) string {
	base := fmt.Sprintf(`arc: %d
title: %s
last-reviewed: %s
`, options.Number, options.Title, now)
	if options.ImplementationRequired {
		base += `reference-implementation:
  status: planned
  notes: ""
`
	}
	base += `adoption:
  wallets: []
  explorers: []
  tooling: []
  infra: []
  dapps-protocols: []
summary:
  adoption-readiness: low
  blockers: []
  notes: ""
`
	return base
}

func renderVettedAdopters() string {
	return `wallets: []
explorers: []
tooling: []
infra: []
dapps-protocols: []
`
}
