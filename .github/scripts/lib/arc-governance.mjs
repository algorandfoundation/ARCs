export const TRACKING_ISSUE_LABEL = "arc-tracking";

export const TEMPLATE_MARKERS = [
  "## ARC",
  "## Canonical Artifacts",
  "## Gate Checklist",
];

export function hasTemplateShape(body) {
  return TEMPLATE_MARKERS.every((marker) => body.includes(marker));
}
