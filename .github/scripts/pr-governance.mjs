import { TRACKING_ISSUE_LABEL, hasTemplateShape } from "./lib/arc-governance.mjs";

const AREA_LABELS = [
  { prefix: "ARCs/", label: "area:arc" },
  { prefix: "adoption/", label: "area:adoption" },
  { prefix: "arckit/", label: "area:arckit" },
  { prefix: ".github/", label: "area:github" },
];

const KIND_LABELS = [
  "kind:draft",
  "kind:review",
  "kind:last-call",
  "kind:final",
  "kind:adoption",
  "kind:editorial",
];

const KIND_SIGNALS = [
  { marker: "## ARC Draft Submission", label: "kind:draft" },
  { marker: "## ARC Transition to Review", label: "kind:review" },
  { marker: "## ARC Transition to Last Call", label: "kind:last-call" },
  { marker: "## ARC Transition to Final", label: "kind:final" },
  { marker: "## ARC Adoption / Implementation Update", label: "kind:adoption" },
  { marker: "## Editorial / Non-Normative ARC Update", label: "kind:editorial" },
];

const GATE_KEYWORDS = [
  "status:",
  "last-call-deadline:",
  "idle-since:",
  "adoption-summary:",
  "implementation-url:",
  "implementation-maintainer:",
];

export async function runProcessCheck({ github, context, core }) {
  const files = await listPullRequestFiles(github, context);
  const arcNumbers = extractRelevantARCNumbers(files);

  if (arcNumbers.length === 0) {
    core.info("No new or gate-relevant ARC changes detected.");
    return;
  }

  const failures = [];
  for (const arcNumber of arcNumbers) {
    const issue = await findTrackingIssue(github, context, arcNumber);
    if (!issue) {
      failures.push(`ARC-${arcNumber} is missing a tracking issue with the ${TRACKING_ISSUE_LABEL} label.`);
      continue;
    }
    if (!hasTemplateShape(issue.body || "")) {
      failures.push(`Tracking issue #${issue.number} for ARC-${arcNumber} does not match the tracking issue template shape.`);
    }
    if (!prBodyReferencesIssue(context.payload.pull_request?.body || "", issue)) {
      failures.push(`Pull request body must explicitly reference tracking issue #${issue.number} for ARC-${arcNumber}.`);
    }
  }

  if (failures.length > 0) {
    core.setFailed(failures.join("\n"));
  }
}

export async function runAutoLabel({ github, context, core }) {
  const files = await listPullRequestFiles(github, context);
  const prNumber = context.payload.pull_request.number;
  const desiredLabels = new Set();

  for (const file of files) {
    for (const mapping of AREA_LABELS) {
      if (file.filename.startsWith(mapping.prefix)) {
        desiredLabels.add(mapping.label);
      }
    }
  }

  const kindLabel = detectKindLabel(context.payload.pull_request?.body || "");
  if (kindLabel) {
    desiredLabels.add(kindLabel);
  }

  const issue = await github.rest.issues.get({
    owner: context.repo.owner,
    repo: context.repo.repo,
    issue_number: prNumber,
  });
  const existingLabels = issue.data.labels.map((label) => typeof label === "string" ? label : label.name);

  const labelsToAdd = [...desiredLabels].filter((label) => !existingLabels.includes(label));
  if (labelsToAdd.length > 0) {
    await github.rest.issues.addLabels({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: prNumber,
      labels: labelsToAdd,
    });
  }

  if (kindLabel) {
    const conflicting = existingLabels.filter((label) => KIND_LABELS.includes(label) && label !== kindLabel);
    for (const label of conflicting) {
      try {
        await github.rest.issues.removeLabel({
          owner: context.repo.owner,
          repo: context.repo.repo,
          issue_number: prNumber,
          name: label,
        });
      } catch (error) {
        core.warning(`Could not remove label ${label}: ${error.message}`);
      }
    }
  } else {
    core.info("No PR template signal found for a kind:* label.");
  }
}

async function listPullRequestFiles(github, context) {
  return github.paginate(github.rest.pulls.listFiles, {
    owner: context.repo.owner,
    repo: context.repo.repo,
    pull_number: context.payload.pull_request.number,
    per_page: 100,
  });
}

function extractRelevantARCNumbers(files) {
  const numbers = new Set();
  for (const file of files) {
    const match = file.filename.match(/^ARCs\/arc-(\d{4})\.md$/);
    if (!match) {
      continue;
    }
    const arcNumber = match[1];
    if (file.status === "added" || gateRelevantPatch(file.patch || "")) {
      numbers.add(arcNumber);
    }
  }
  return [...numbers].sort();
}

function gateRelevantPatch(patch) {
  return GATE_KEYWORDS.some((keyword) => patch.includes(keyword));
}

async function findTrackingIssue(github, context, arcNumber) {
  const query = `repo:${context.repo.owner}/${context.repo.repo} is:issue label:${TRACKING_ISSUE_LABEL} "ARC-${arcNumber}"`;
  const result = await github.rest.search.issuesAndPullRequests({
    q: query,
    per_page: 20,
  });
  return result.data.items.find((item) => !item.pull_request) || null;
}

function prBodyReferencesIssue(body, issue) {
  return body.includes(`#${issue.number}`) || body.includes(issue.html_url);
}

function detectKindLabel(body) {
  for (const signal of KIND_SIGNALS) {
    if (body.includes(signal.marker)) {
      return signal.label;
    }
  }
  return "";
}
