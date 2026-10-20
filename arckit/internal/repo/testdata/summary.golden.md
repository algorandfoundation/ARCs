# ARC State Summary

## Validation Snapshot

- Generated: `2026-04-10T09:30:00Z`
- Repository: `/repo`
- Validation summary: `3 error(s), 0 warning(s), 0 info`

## State Overview

- Total ARCs: `6`
- ARCs with assets: `1`
- Adoption summary coverage: `5/6`
- Total adoption files: `5`
- Total adopter entries: `5`
- ARCs with at least one adopter: `4`

### Counts by Status

| Status | Count |
| --- | --- |
| Draft | 1 |
| Last Call | 1 |
| Final | 3 |
| Idle | 1 |

### Counts by Type

| Type | Count |
| --- | --- |
| Standards Track | 5 |
| Meta | 1 |

### Adoption Readiness Distribution

| Adoption Readiness | Count |
| --- | --- |
| low | 5 |

### Reference Implementation Status Counts

| Reference Implementation Status | Count |
| --- | --- |
| wip | 1 |
| shipped | 1 |

## Transition Watch

### Overdue Last Call

| ARC | Title | Deadline | Days Overdue | Editor Action |
| --- | --- | --- | --- | --- |
| 44 | Last Call ARC | 2026-04-01 | 9 | Decide Final, extend Last Call, or move back to Review |

### Upcoming Last Call (next 14 days)

None

### Idle ARCs

| ARC | Title | Idle Since | Days Idle | Editor Action |
| --- | --- | --- | --- | --- |
| 47 | Idle ARC With WIP Implementation | 2026-02-01 | 68 | Confirm Idle should remain or restart editorial follow-up |

### Implementation-Required ARCs Not Shipped

| ARC | Title | Status | Ref Impl Status | Editor Action |
| --- | --- | --- | --- | --- |
| 47 | Idle ARC With WIP Implementation | Idle | wip | Check canonical implementation readiness before transition |

## Adoption Watch

### Final ARCs With Zero Adopters

| ARC | Title | Adoption Readiness | Last Reviewed | Editor Action |
| --- | --- | --- | --- | --- |
| 45 | Final ARC With Placeholder Adoption | low | 2026-03-01 | Backfill at least one tracked adopter or confirm historical exception |

### Final ARCs With 1-2 Adopters

| ARC | Title | Adoption Readiness | Last Reviewed | Editor Action |
| --- | --- | --- | --- | --- |
| 48 | Final ARC With One Adopter | low | 2026-04-05 | Check whether additional adoption evidence exists |
| 46 | Final ARC With Two Adopters | low | 2026-04-05 | Check whether additional adoption evidence exists |

### Stale Adoption Reviews (>30 days)

| ARC | Title | Status | Last Reviewed | Age (days) | Editor Action |
| --- | --- | --- | --- | --- | --- |
| 44 | Last Call ARC | Last Call | 2026-03-01 | 40 | Refresh adoption summary and evidence |
| 45 | Final ARC With Placeholder Adoption | Final | 2026-03-01 | 40 | Refresh adoption summary and evidence |
| 47 | Idle ARC With WIP Implementation | Idle | 2026-03-05 | 36 | Refresh adoption summary and evidence |

### Adoption Totals

#### Adopter Entries by Category

| Category | Count |
| --- | --- |
| wallets | 2 |
| explorers | 1 |
| tooling | 2 |
| infra | 0 |
| dapps-protocols | 0 |

#### Adopter Entries by Actor Status

| Actor Status | Count |
| --- | --- |
| planned | 0 |
| in_progress | 1 |
| shipped | 4 |
| declined | 0 |
| unknown | 0 |

#### Top Adopters by Distinct ARC Coverage

| Adopter | Distinct ARCs |
| --- | --- |
| tool-one | 2 |
| wallet-one | 2 |
| explorer-one | 1 |

#### Top 10 ARCs by Adopter Count

| ARC | Title | Status | Adopters |
| --- | --- | --- | --- |
| 46 | Final ARC With Two Adopters | Final | 2 |
| 44 | Last Call ARC | Last Call | 1 |
| 47 | Idle ARC With WIP Implementation | Idle | 1 |
| 48 | Final ARC With One Adopter | Final | 1 |

_Note: This table is limited to 10 rows. Ties are broken by lower ARC number after sorting by adopter count._

## Relationship Watch

### Top ARCs Most Referenced by `requires`

| ARC | Title | Requires References |
| --- | --- | --- |
| 44 | Last Call ARC | 2 |
| 46 | Final ARC With Two Adopters | 1 |

### Top ARCs Most Referenced by `extends`

| ARC | Title | Extends References |
| --- | --- | --- |
| 46 | Final ARC With Two Adopters | 2 |

### Non-Empty `supersedes` / `superseded-by` Pairs

| ARC | Title | Supersedes | Superseded By |
| --- | --- | --- | --- |
| 45 | Final ARC With Placeholder Adoption | None | 46 |
| 46 | Final ARC With Two Adopters | 45 | None |

## Data Notes

- This report is local and offline-only.
- It uses ARC front matter, adoption YAML, vetted adopters, asset directories, and ARC relationship fields.
- State age is only known where explicit dates exist: `last-call-deadline`, `idle-since`, and `last-reviewed`.
- Missing `sponsor`, missing `implementation-required`, and sparse `updated` usage are migration-state and are not treated as backlog in this report.
