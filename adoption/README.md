# ARC Adoption Summaries

This directory contains machine-readable adoption summaries for numbered ARCs.

They are operational artifacts used to record:

- implementation ownership,
- implementation status,
- ecosystem adoption signals,
- evidence used for status transitions.

This directory also contains the canonical vetted adopters registry:

- `vetted-adopters.yaml`

Every adopter `name` used in an ARC adoption summary must:

- be lower-kebab-case;
- appear in the matching category of `vetted-adopters.yaml`.

When an ARC adoption summary is in `Final` status, at least one of its adoption
categories must contain at least one adopter entry. A `Final` summary with all
adoption categories empty is invalid.

An ARC expected to be adopted by ecosystem implementers should keep this file current
throughout `Draft`, `Review`, `Last Call`, `Final`, `Idle`, and `Deprecated` states.
