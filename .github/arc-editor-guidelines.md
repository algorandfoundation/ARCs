# ARC Editor Guidelines

This document describes the ARC Editor workflow as it is currently automated in this
repository. It focuses on what GitHub Actions and repository policy already enforce,
and what still requires a human action on GitHub.

## Source of Truth

Use these documents together:

- [`ARCs/arc-0000.md`](../ARCs/arc-0000.md) for the ARC process and editor responsibilities;
- [`README.md`](../README.md) for the canonical venues and PR templates;
- [`.github/ci-cd-release-specs.md`](./ci-cd-release-specs.md) for the current CI/CD policy;
- [`.github/ISSUE_TEMPLATE/arc-tracking.md`](./ISSUE_TEMPLATE/arc-tracking.md) for the tracking issue shape.

## What Is Automated Today

### Pull request validation

The CI/CD spec defines a stable check surface for pull requests. In the current
implementation, GitHub Actions runs:

- `hygiene` from the repository-root `pre-commit` config for merge markers, line endings, whitespace/newline policy, YAML syntax/formatting, and generic Markdown linting;
- `repo-validate-offline` for the canonical offline repository gate;
- `arc-process-check` for GitHub-native process enforcement;
- `arckit-tool` when `arckit/**` or workflow files change;

`arckit` is limited to ARC-specific validation. Generic Markdown, YAML, and
text-file hygiene is intentionally handled outside the CLI.

The offline validation gate implicitly applies the repository-root `.arckit.jsonc`
when present. An invalid `.arckit.jsonc` blocks the pull request in the same way as
any other validation failure.

These checks can block a pull request, but they do not make editorial decisions.

### ARC process checks

For new numbered ARC pull requests, or pull requests that change gate-relevant ARC
metadata, automation currently checks that:

- a matching tracking issue exists;
- the tracking issue matches the repository tracking issue template shape;
- the tracking issue carries the `arc-tracking` label;
- the PR body explicitly references the tracking issue.

Automation fails the PR if any of those repository-process requirements are missing.

### Automatic PR labels

The `PR governance` workflow currently applies labels automatically:

- `area:*` labels from changed paths;
- one `kind:*` label from the selected PR template body markers.

If the PR body does not contain a recognized template marker, no `kind:*` label is
inferred automatically.

The CI/CD spec also defines the supported label vocabulary for editor workflows,
including `arc-tracking`, `maintenance-report`, `maintenance:authors-action`, and
`maintenance:editor-action`. Of those, the current automation actively applies
`area:*`, `kind:*`, and `maintenance-report`; the `maintenance:*` action labels are
policy vocabulary, but are not currently applied by the monthly workflow.

### Monthly maintenance report

There is a `Monthly ARC maintenance` workflow that can:

- run offline validation;
- run online validation;
- detect missing or malformed tracking issues;
- detect inactivity that may suggest `Stagnant` or `Idle`;
- create or update one rollup GitHub issue for the current month with author and editor actions.

This workflow runs on a monthly cron schedule and can also be started manually.
It always writes a workflow summary. It creates or updates the monthly rollup issue
only when the run finds actionable maintenance or online-link findings.

### Editor reminders from ARC-Kit

`arckit` can validate machine-checkable transition requirements, but it only emits
informational reminders for editorial checks such as:

- consensus and dissent handling;
- adequacy of adoption evidence;
- quality and maintenance of the reference implementation;
- explicit editor approval.

Passing `arckit` validation is therefore necessary, but not sufficient, for an editor
decision.

## What Still Requires Manual GitHub Action

The following ARC editor responsibilities are not automated today and still require
manual action in GitHub.

### Pre-ARC and numbering

An editor must still manually:

- review the `Pre-ARC` discussion;
- decide whether the proposal should become an ARC;
- close or steer the discussion with the appropriate outcome;
- assign the ARC number.

No workflow creates, closes, or adjudicates `Pre-ARC` discussions.

### Tracking issue creation and upkeep

An editor must still manually:

- confirm that the numbered ARC tracking issue exists;
- ensure the issue uses the tracking issue template;
- keep the gate checklist and decision log meaningful and current when editorial decisions are made;
- verify that the canonical links in the issue point to the correct discussion, PR, adoption summary, and implementation repository;
- confirm that any newly named adopters are present in `adoption/vetted-adopters.yaml`;
- confirm that the ARC front matter, not the adoption summary, carries the canonical `status`, `sponsor`, `implementation-required`, implementation URL, and implementation maintainers;
- confirm that any ARC moving to or remaining in `Final` still has at least one tracked adopter in its adoption summary.

The current CI/CD spec treats tracking issue creation as a manual author action after
number assignment. From the editor side, the manual responsibility is to confirm that
the issue exists and is correct. Automation verifies presence and basic shape; it
does not create the issue, fill in its contents, or decide whether the issue is
substantively current.

### Editorial review and transition approval

An editor must still manually:

- review the ARC content for scope, clarity, and process readiness;
- determine whether unresolved normative questions remain;
- judge whether adoption evidence is credible and adequate;
- confirm that a `Final` ARC does not have an empty adoption summary;
- confirm that implementation identity changes are recorded in ARC front matter while implementation status changes are recorded in the adoption summary;
- judge whether the reference implementation is sufficiently maintained;
- confirm editor approval for status transitions;
- request changes, approve, merge, or close the PR.

Automation validates machine-verifiable prerequisites. It does not approve a transition
or replace editorial judgment.

### Status management

An editor must still manually manage or approve GitHub changes related to:

- transition to `Review`;
- transition to `Last Call`, including setting `last-call-deadline`;
- transition to `Final`;
- follow-up on `Stagnant`;
- editor review for `Idle`;
- `Withdrawn`, `Deprecated`, and `Living` outcomes where applicable.

Automation may suggest or validate parts of these transitions, but it does not mutate
ARC status on its own.

### Label correction and exception handling

An editor must still manually correct labels when:

- the PR body does not match a supported template marker;
- the inferred `kind:*` label is wrong for the actual review stage;
- exceptional repository maintenance requires a deliberate manual label choice.

### Monthly maintenance triage

An editor must still manually:

- review the generated workflow summary or rollup issue;
- carry out the resulting follow-up on issues, PRs, and status decisions.

The workflow reports findings. It does not resolve them.

## Practical Rule

Treat automation as a gate and reminder system:

- if a workflow fails, fix the deterministic repository problem first;
- if all workflows pass, perform the remaining editorial review in GitHub before merging;
- if the monthly report suggests action, record the human decision in the relevant tracking issue or pull request.

## Current Non-Goals

The repository does not currently automate, and the CI/CD spec explicitly keeps out
of scope:

- creating tracking issues;
- assigning ARC numbers;
- closing `Pre-ARC` discussions;
- changing ARC status automatically;
- recording editor approval automatically;
- resolving consensus or dissent questions automatically.
