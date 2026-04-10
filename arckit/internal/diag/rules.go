package diag

var rules = []Rule{
	{
		ID:          "R:001",
		Severity:    SeverityError,
		Title:       "Missing front matter",
		Description: "ARC files must start with a YAML front matter block.",
		Rationale:   "The repository depends on deterministic metadata extraction.",
		Hint:        "Add a front matter block delimited by --- at the top of the file.",
	},
	{
		ID:          "R:002",
		Severity:    SeverityError,
		Title:       "Invalid ARC filename",
		Description: "ARC files must live under ARCs/ and use arc-####.md.",
		Rationale:   "Repository-wide indexing depends on stable paths.",
		Hint:        "Rename the file to ARCs/arc-####.md.",
	},
	{
		ID:          "R:003",
		Severity:    SeverityError,
		Title:       "Invalid front matter order",
		Description: "Known ARC front matter keys must appear in the canonical order.",
		Rationale:   "Stable ordering keeps formatting and validation deterministic.",
		Hint:        "Reorder the front matter keys to match the spec.",
	},
	{
		ID:          "R:004",
		Severity:    SeverityError,
		Title:       "Missing required ARC field",
		Description: "A required ARC front matter field is missing or empty.",
		Rationale:   "Core ARC metadata must be complete for tooling and review.",
		Hint:        "Add the required field with a valid value.",
	},
	{
		ID:          "R:005",
		Severity:    SeverityError,
		Title:       "Invalid front matter",
		Description: "ARC front matter must be valid YAML.",
		Rationale:   "Invalid YAML prevents deterministic parsing.",
		Hint:        "Fix the YAML syntax in the front matter block.",
	},
	{
		ID:          "R:006",
		Severity:    SeverityError,
		Title:       "Unknown ARC field",
		Description: "Unknown ARC front matter fields are not allowed in v1.",
		Rationale:   "v1 intentionally keeps the metadata surface fixed.",
		Hint:        "Remove the unknown field or move the data into the document body.",
	},
	{
		ID:          "R:007",
		Severity:    SeverityError,
		Title:       "Invalid ARC field value",
		Description: "An ARC field has the wrong type or an unsupported value.",
		Rationale:   "Field semantics must be machine-verifiable.",
		Hint:        "Use the type and enum values required by the specification.",
	},
	{
		ID:          "R:008",
		Severity:    SeverityError,
		Title:       "Missing required ARC section",
		Description: "Every ARC must contain the required level-2 sections.",
		Rationale:   "Stable sections are part of the ARC review contract.",
		Hint:        "Add the missing section with a level-2 heading.",
	},
	{
		ID:          "R:009",
		Severity:    SeverityError,
		Title:       "Invalid local link",
		Description: "A repo-local link is invalid, root-relative, or points to a missing target.",
		Rationale:   "Offline validation must keep repository references consistent.",
		Hint:        "Use a relative path that resolves to an existing file in the repository.",
	},
	{
		ID:          "R:010",
		Severity:    SeverityError,
		Title:       "Invalid asset link",
		Description: "Asset links from an ARC must stay inside the matching assets/arc-#### subtree.",
		Rationale:   "ARC assets are scoped to the ARC that owns them.",
		Hint:        "Move the asset or update the link to the matching asset directory.",
	},
	{
		ID:          "R:011",
		Severity:    SeverityError,
		Title:       "Invalid ARC relationship",
		Description: "An ARC relationship field is invalid or inconsistent.",
		Rationale:   "Cross-ARC relationships must stay machine-readable.",
		Hint:        "Use numeric ARC identifiers and keep reciprocal fields aligned.",
	},
	{
		ID:          "R:012",
		Severity:    SeverityError,
		Title:       "Adoption summary required",
		Description: "This ARC status requires an adoption summary.",
		Rationale:   "Status transitions and maintenance depend on adoption evidence.",
		Hint:        "Set adoption-summary to the matching file under adoption/ and add the file.",
	},
	{
		ID:          "R:013",
		Severity:    SeverityError,
		Title:       "Invalid adoption summary reference",
		Description: "The adoption-summary field must point to the ARC's matching adoption file.",
		Rationale:   "ARC-to-adoption mapping must stay deterministic.",
		Hint:        "Point adoption-summary to adoption/arc-####.yaml.",
	},
	{
		ID:          "R:014",
		Severity:    SeverityError,
		Title:       "Invalid adoption filename",
		Description: "Adoption summaries must live under adoption/ and use arc-####.yaml.",
		Rationale:   "Repository-wide discovery depends on stable naming.",
		Hint:        "Rename the file to adoption/arc-####.yaml.",
	},
	{
		ID:          "R:015",
		Severity:    SeverityError,
		Title:       "Missing required adoption field",
		Description: "A required adoption-summary field is missing.",
		Rationale:   "Adoption summaries have a fixed minimum schema in v1.",
		Hint:        "Add the missing field to the YAML file.",
	},
	{
		ID:          "R:016",
		Severity:    SeverityError,
		Title:       "Invalid adoption field",
		Description: "An adoption-summary field has the wrong type or an unsupported value.",
		Rationale:   "The adoption summary must stay machine-readable.",
		Hint:        "Use the type and enum values defined in the specification.",
	},
	{
		ID:          "R:017",
		Severity:    SeverityError,
		Title:       "Adoption summary mismatch",
		Description: "The adoption summary conflicts with the corresponding ARC identity or file mapping.",
		Rationale:   "ARC and adoption files must stay aligned on which ARC number the summary belongs to.",
		Hint:        "Update the ARC number or file mapping so the summary matches the corresponding ARC.",
	},
	{
		ID:          "R:018",
		Severity:    SeverityError,
		Title:       "Repository mapping conflict",
		Description: "Repository-wide ARC, adoption, or asset discovery found a conflict.",
		Rationale:   "The repository must have one consistent file tree for each ARC.",
		Hint:        "Remove duplicates or add the missing canonical file.",
	},
	{
		ID:          "R:019",
		Severity:    SeverityError,
		Title:       "Transition requirement failed",
		Description: "The requested machine-verifiable status transition requirement was not met.",
		Rationale:   "Transition checks guard the ARC process before editor approval.",
		Hint:        "Add the missing metadata, evidence, or sections required for the target status.",
	},
	{
		ID:          "R:020",
		Severity:    SeverityInfo,
		Title:       "Manual transition reminder",
		Description: "Some transition checks remain editorial and must be confirmed by humans.",
		Rationale:   "The CLI only validates the machine-verifiable subset of transitions.",
		Hint:        "Confirm the remaining editorial checks in the tracking issue and review process.",
	},
	{
		ID:          "R:021",
		Severity:    SeverityError,
		Title:       "Non-canonical ARC YAML field shape",
		Description: "List-valued ARC front matter fields must use canonical YAML sequences.",
		Rationale:   "Canonical YAML-native field shapes avoid ambiguous comma-separated encodings and keep tooling deterministic.",
		Hint:        "Use YAML sequences for list-valued ARC fields such as author, updated, implementation-maintainer, requires, supersedes, extends, and extended-by.",
	},
	{
		ID:          "R:022",
		Severity:    SeverityError,
		Title:       "Invalid vetted adopters registry",
		Description: "The canonical vetted adopters registry is missing or does not match the required schema.",
		Rationale:   "Adoption evidence must refer to a repository-scoped vetted adopter list to keep actor names deterministic.",
		Hint:        "Create or fix adoption/vetted-adopters.yaml using the required category keys and lower-kebab adopter names.",
	},
	{
		ID:          "R:023",
		Severity:    SeverityError,
		Title:       "Invalid adopter reference",
		Description: "An adoption actor name is not lower-kebab-case or is not present in the vetted adopters registry.",
		Rationale:   "Per-ARC adoption entries must use canonical actor identifiers drawn from the vetted registry.",
		Hint:        "Use a lower-kebab adopter name from adoption/vetted-adopters.yaml in the matching category.",
	},
	{
		ID:          "R:024",
		Severity:    SeverityError,
		Title:       "Blank line in front matter",
		Description: "ARC front matter must not contain empty lines.",
		Rationale:   "Blank lines inside the front matter block make canonical formatting and line-based tooling less deterministic.",
		Hint:        "Remove empty lines from the front matter block or run arckit fmt.",
	},
	{
		ID:          "R:025",
		Severity:    SeverityError,
		Title:       "Final ARC has no tracked adoption",
		Description: "An ARC in Final status must not have all adoption categories empty.",
		Rationale:   "Final ARCs must record at least one concrete ecosystem adopter so the repository does not treat empty placeholder summaries as real adoption.",
		Hint:        "Add at least one vetted adopter entry to the adoption summary before marking the ARC Final.",
	},
	{
		ID:          "R:026",
		Severity:    SeverityError,
		Title:       "Invalid invocation",
		Description: "The command invocation is missing required arguments or uses unsupported values.",
		Rationale:   "Stable exit codes depend on distinguishing usage errors from validation failures.",
		Hint:        "Check the command help and provide the required arguments.",
	},
	{
		ID:          "R:027",
		Severity:    SeverityError,
		Title:       "Runtime failure",
		Description: "A runtime failure prevented normal validation or file generation.",
		Rationale:   "Unexpected runtime problems should be reported distinctly from semantic validation results.",
		Hint:        "Check file paths, permissions, and tool availability, then retry.",
	},
	{
		ID:          "R:028",
		Severity:    SeverityError,
		Title:       "Invalid arckit configuration",
		Description: "The repository .arckit.jsonc file could not be parsed or did not match the supported schema.",
		Rationale:   "Repository-local suppressions must be valid so CI and local validation stay deterministic.",
		Hint:        "Fix the .arckit.jsonc syntax, selectors, or rule IDs, then retry.",
	},
	{
		ID:          "R:029",
		Severity:    SeverityError,
		Title:       "Non-canonical implementation URL",
		Description: "An implementation-required ARC must declare the canonical sponsor-specific GitHub repository URL in front matter.",
		Rationale:   "Canonical implementation repository URLs keep implementation ownership deterministic and machine-verifiable across the ARC process.",
		Hint:        "Set implementation-url to the exact sponsor-specific GitHub repository path for the ARC number.",
	},
}

var ruleIndex = func() map[string]Rule {
	index := make(map[string]Rule, len(rules))
	for _, rule := range rules {
		index[rule.ID] = rule
	}
	return index
}()

func AllRules() []Rule {
	out := make([]Rule, len(rules))
	copy(out, rules)
	return out
}

func Lookup(id string) (Rule, bool) {
	rule, ok := ruleIndex[id]
	return rule, ok
}
