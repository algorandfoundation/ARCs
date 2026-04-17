import test from "node:test";
import assert from "node:assert/strict";

import { parseArcMetadataText } from "../lib/arc-metadata.mjs";

test("parseArcMetadataText preserves current scalar and sequence behavior", () => {
  const metadata = parseArcMetadataText(`---
arc: 0042
title: Example ARC
status: Review
author:
  - @alice
  - @bob
adoption-summary: adoption/arc-0042.yaml
---

# Body
`);

  assert.deepEqual(metadata, {
    arc: "0042",
    title: "Example ARC",
    status: "Review",
    author: "@alice, @bob",
    adoptionSummary: "adoption/arc-0042.yaml",
  });
});

test("parseArcMetadataText returns empty fields when front matter is absent", () => {
  assert.deepEqual(parseArcMetadataText("# No front matter\n"), {
    arc: "",
    title: "",
    status: "",
    author: "",
    adoptionSummary: "",
  });
});
