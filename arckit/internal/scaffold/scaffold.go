package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
)

type InitOptions struct {
	Root                   string
	Number                 int
	Title                  string
	Type                   string
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
	created := []string{arcPath, adoptionPath, assetPath}

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
		author = "TBD"
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
	return created, nil, nil
}

func renderARC(options InitOptions, author string, description string, now string) string {
	number := fmt.Sprintf("%04d", options.Number)
	return fmt.Sprintf(`---
arc: %d
title: %s
description: %s
author: %s
discussions-to:
status: Draft
type: %s
created: %s
sponsor: %s
implementation-required: %t
adoption-summary: adoption/arc-%s.yaml
---

## Abstract

TODO: add the abstract.

## Motivation

TODO: explain the problem or opportunity.

## Specification

TODO: define the proposed behavior.

## Rationale

TODO: explain the design tradeoffs.

## Security Considerations

TODO: document the relevant security considerations.

`, options.Number, options.Title, description, author, options.Type, now, options.Sponsor, options.ImplementationRequired, number)
}

func renderAdoption(options InitOptions, now string) string {
	base := fmt.Sprintf(`arc: %d
title: %s
status: Draft
last-reviewed: %s
sponsor: %s
implementation-required: %t
`, options.Number, options.Title, now, options.Sponsor, options.ImplementationRequired)
	if options.ImplementationRequired {
		base += `reference-implementation:
  repository: ""
  maintainers: []
  status: planned
  notes: ""
`
	}
	base += `adoption:
  wallets: []
  explorers: []
  sdk-libraries: []
  infra: []
  dapps-protocols: []
summary:
  adoption-readiness: low
  blockers: []
  notes: ""
`
	return base
}
