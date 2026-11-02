export const TRACKING_ISSUE_LABEL = "arc-tracking";

export const TEMPLATE_MARKERS = [
  "## ARC",
  "## Canonical Artifacts",
  "## Gate Checklist",
];

export function hasTemplateShape(body) {
  return TEMPLATE_MARKERS.every((marker) => body.includes(marker));
}

export const KIND_LABELS = [
  "kind:draft",
  "kind:review",
  "kind:last-call",
  "kind:final",
  "kind:adoption",
  "kind:editorial",
];

const ARC_FILE_PATTERN = /^ARCs\/arc-\d{4}\.md$/;

const STATUS_KIND_LABELS = {
  "Review": "kind:review",
  "Last Call": "kind:last-call",
  "Final": "kind:final",
};

// Derives the PR kind from its changed files (GitHub pulls.listFiles entries), most
// significant first: new ARC, status change, adoption update, then other ARC edits.
export function detectKindLabel(files) {
  const arcFiles = files.filter((file) => ARC_FILE_PATTERN.test(file.filename));
  if (arcFiles.some((file) => file.status === "added")) {
    return "kind:draft";
  }
  for (const file of arcFiles) {
    const label = STATUS_KIND_LABELS[changedStatus(file.patch || "")];
    if (label) {
      return label;
    }
  }
  if (files.some((file) => file.filename.startsWith("adoption/"))) {
    return "kind:adoption";
  }
  if (arcFiles.length > 0) {
    return "kind:editorial";
  }
  return "";
}

function changedStatus(patch) {
  const removed = patch.match(/^-status:\s*(.+?)\s*$/m);
  const added = patch.match(/^\+status:\s*(.+?)\s*$/m);
  if (!removed || !added || removed[1] === added[1]) {
    return "";
  }
  return added[1];
}
