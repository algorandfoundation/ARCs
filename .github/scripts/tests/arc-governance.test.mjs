import test from "node:test";
import assert from "node:assert/strict";

import { TEMPLATE_MARKERS, detectKindLabel, hasTemplateShape } from "../lib/arc-governance.mjs";

test("hasTemplateShape accepts bodies with all required markers", () => {
  const body = `${TEMPLATE_MARKERS[0]}\n...\n${TEMPLATE_MARKERS[1]}\n...\n${TEMPLATE_MARKERS[2]}`;
  assert.equal(hasTemplateShape(body), true);
});

test("hasTemplateShape rejects bodies missing any required marker", () => {
  const body = `${TEMPLATE_MARKERS[0]}\n...\n${TEMPLATE_MARKERS[1]}`;
  assert.equal(hasTemplateShape(body), false);
});

test("detectKindLabel labels a new ARC file as draft", () => {
  assert.equal(detectKindLabel([{ filename: "ARCs/arc-0099.md", status: "added" }]), "kind:draft");
});

test("detectKindLabel labels a status change by its target status", () => {
  const patch = "@@ -7,1 +7,1 @@\n-status: Last Call\n+status: Final\n";
  assert.equal(detectKindLabel([{ filename: "ARCs/arc-0042.md", status: "modified", patch }]), "kind:final");
});

test("detectKindLabel ignores status changes without a gated target", () => {
  const patch = "@@ -7,1 +7,1 @@\n-status: Draft\n+status: Stagnant\n";
  assert.equal(detectKindLabel([{ filename: "ARCs/arc-0042.md", status: "modified", patch }]), "kind:editorial");
});

test("detectKindLabel labels adoption-only changes as adoption", () => {
  assert.equal(detectKindLabel([{ filename: "adoption/arc-0042.yaml", status: "modified", patch: "" }]), "kind:adoption");
});

test("detectKindLabel labels other ARC edits as editorial", () => {
  assert.equal(detectKindLabel([{ filename: "ARCs/arc-0042.md", status: "modified", patch: "+typo fix\n" }]), "kind:editorial");
});

test("detectKindLabel returns no label for non-ARC changes", () => {
  assert.equal(detectKindLabel([{ filename: ".github/workflows/pr-validation.yml", status: "modified" }]), "");
});
