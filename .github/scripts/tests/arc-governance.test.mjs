import test from "node:test";
import assert from "node:assert/strict";

import { TEMPLATE_MARKERS, hasTemplateShape } from "../lib/arc-governance.mjs";

test("hasTemplateShape accepts bodies with all required markers", () => {
  const body = `${TEMPLATE_MARKERS[0]}\n...\n${TEMPLATE_MARKERS[1]}\n...\n${TEMPLATE_MARKERS[2]}`;
  assert.equal(hasTemplateShape(body), true);
});

test("hasTemplateShape rejects bodies missing any required marker", () => {
  const body = `${TEMPLATE_MARKERS[0]}\n...\n${TEMPLATE_MARKERS[1]}`;
  assert.equal(hasTemplateShape(body), false);
});
