<!-- markdownlint-disable MD041 -->

## ARC Status Change

Use this template when requesting a status change of an ARC (for example `Draft` to
`Review`, `Review` to `Last Call`, or `Last Call` to `Final`).

## Summary

Describe why the ARC is ready for the requested status.

## Canonical References

- Tracking Issue:
- ARC File:
- Adoption Summary File (required from `Last Call`):
- Reference Implementation Repository:
- Related Issues / PRs:

## Checklist

Machine-checkable transition gates run in CI. Tick only the items for the requested
status.

To `Review`:

- [ ] If `implementation-required: true`, the reference implementation repository exists.

To `Last Call`:

- [ ] If `implementation-required: true`, runnable code, tests, examples, or conformance artifacts exist.

To `Final`:

- [ ] Known interoperability concerns are recorded.

## Notes for Reviewers

Highlight open questions, ecosystem dependencies, or follow-up items.
