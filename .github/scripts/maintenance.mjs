import fs from "node:fs/promises";
import path from "node:path";
import { execFile } from "node:child_process";
import { promisify } from "node:util";

const execFileAsync = promisify(execFile);
const SIX_MONTHS_MS = 180 * 24 * 60 * 60 * 1000;
const STAGNANT_FOLLOW_UP_MS = 210 * 24 * 60 * 60 * 1000;
const ISSUE_TITLE_PREFIX = "Monthly ARC maintenance report";
const ISSUE_LABEL = "maintenance-report";
const TEMPLATE_MARKERS = [
  "## ARC",
  "## Canonical Artifacts",
  "## Gate Checklist",
];

export async function runMaintenance({ github, context, core }) {
  const offlineReport = await loadJSON(process.env.OFFLINE_REPORT_PATH);
  const onlineReport = await loadJSON(process.env.ONLINE_REPORT_PATH);
  const summaryReport = await loadText(process.env.SUMMARY_REPORT_PATH);
  const summaryArtifactUrl = process.env.SUMMARY_ARTIFACT_URL || "";
  const repoRoot = process.cwd();
  const arcDir = path.join(repoRoot, "ARCs");
  const arcFiles = (await fs.readdir(arcDir)).filter((entry) => /^arc-\d{4}\.md$/.test(entry)).sort();

  const authorsAction = [];
  const editorAction = [];
  const suggestedTransitions = [];

  for (const fileName of arcFiles) {
    const filePath = path.join("ARCs", fileName);
    const metadata = await parseArcMetadata(path.join(repoRoot, filePath));
    if (!metadata.arc) {
      continue;
    }
    const trackingIssue = await findTrackingIssue(github, context, metadata.arc);
    if (!trackingIssue) {
      editorAction.push(`- ARC-${metadata.arc} ${metadata.title || ""}: create the missing tracking issue with the \`arc-tracking\` label.`);
    } else if (!hasTemplateShape(trackingIssue.body || "")) {
      editorAction.push(`- ARC-${metadata.arc} ${metadata.title || ""}: tracking issue #${trackingIssue.number} does not match the expected template shape.`);
    }

    const latestActivity = await latestCanonicalActivity({
      github,
      context,
      arcNumber: metadata.arc,
      trackingIssue,
      arcPath: filePath,
      adoptionPath: metadata.adoptionSummary,
    });
    if (!latestActivity) {
      continue;
    }

    const age = Date.now() - latestActivity.getTime();
    if (["Draft", "Review", "Last Call"].includes(metadata.status) && age >= SIX_MONTHS_MS) {
      suggestedTransitions.push(`- ARC-${metadata.arc} ${metadata.title || ""}: suggest \`Stagnant\` due to no canonical activity since ${latestActivity.toISOString().slice(0, 10)}.${authorMentions(metadata.author)}`);
    }
    if (metadata.status === "Final" && age >= SIX_MONTHS_MS) {
      suggestedTransitions.push(`- ARC-${metadata.arc} ${metadata.title || ""}: suggest editor review for \`Idle\` due to no canonical activity since ${latestActivity.toISOString().slice(0, 10)}.${authorMentions(metadata.author)}`);
    }
    if (metadata.status === "Stagnant" && age >= STAGNANT_FOLLOW_UP_MS) {
      editorAction.push(`- ARC-${metadata.arc} ${metadata.title || ""}: follow up on existing \`Stagnant\` ARC with no new activity since ${latestActivity.toISOString().slice(0, 10)}.`);
    }
  }

  for (const diagnostic of (offlineReport.diagnostics || []).filter((item) => item.severity === "error")) {
    authorsAction.push(`- ${formatDiagnostic(diagnostic)}.`);
  }

  for (const diagnostic of (onlineReport.diagnostics || []).filter((item) => item.severity !== "info")) {
    authorsAction.push(`- Advisory online finding: ${formatDiagnostic(diagnostic)}.`);
  }

  const month = new Date().toISOString().slice(0, 7);
  const body = renderReport(month, authorsAction, editorAction, suggestedTransitions);

  await core.summary
    .addHeading(`ARC maintenance report ${month}`)
    .addRaw(body, true)
    .addRaw(renderRepoStateSummary(summaryReport, summaryArtifactUrl), true)
    .write();

  if (authorsAction.length === 0 && editorAction.length === 0 && suggestedTransitions.length === 0) {
    core.info("No actionable monthly maintenance findings.");
    return;
  }

  await upsertMonthlyIssue(github, context, month, body);
}

async function loadJSON(filePath) {
  if (!filePath) {
    return {};
  }
  try {
    const content = await fs.readFile(filePath, "utf8");
    return JSON.parse(content);
  } catch {
    return {};
  }
}

async function loadText(filePath) {
  if (!filePath) {
    return "";
  }
  try {
    return await fs.readFile(filePath, "utf8");
  } catch {
    return "";
  }
}

async function parseArcMetadata(filePath) {
  const content = await fs.readFile(filePath, "utf8");
  const match = content.match(/^---\n([\s\S]*?)\n---/);
  const frontMatter = match ? match[1] : "";
  const fields = parseTopLevelFrontMatter(frontMatter);
  return {
    arc: scalarField(fields, "arc"),
    title: scalarField(fields, "title"),
    status: scalarField(fields, "status"),
    author: listOrScalarField(fields, "author"),
    adoptionSummary: scalarField(fields, "adoption-summary"),
  };
}

function parseTopLevelFrontMatter(frontMatter) {
  const fields = {};
  let currentKey = "";

  for (const line of frontMatter.split(/\r?\n/)) {
    if (!line.trim()) {
      continue;
    }
    const keyMatch = line.match(/^([A-Za-z0-9-]+):(?:\s*(.*))?$/);
    if (keyMatch) {
      currentKey = keyMatch[1];
      const value = (keyMatch[2] || "").trim();
      fields[currentKey] = value ? value : [];
      continue;
    }
    const sequenceMatch = line.match(/^\s*-\s*(.+)$/);
    if (sequenceMatch && currentKey && Array.isArray(fields[currentKey])) {
      fields[currentKey].push(sequenceMatch[1].trim());
    }
  }

  return fields;
}

function scalarField(fields, key) {
  const value = fields[key];
  if (Array.isArray(value)) {
    return value[0] || "";
  }
  return value || "";
}

function listOrScalarField(fields, key) {
  const value = fields[key];
  if (Array.isArray(value)) {
    return value.join(", ");
  }
  return value || "";
}

async function findTrackingIssue(github, context, arcNumber) {
  const query = `repo:${context.repo.owner}/${context.repo.repo} is:issue label:arc-tracking "ARC-${arcNumber.toString().padStart(4, "0")}"`;
  const result = await github.rest.search.issuesAndPullRequests({
    q: query,
    per_page: 20,
  });
  return result.data.items.find((item) => !item.pull_request) || null;
}

function hasTemplateShape(body) {
  return TEMPLATE_MARKERS.every((marker) => body.includes(marker));
}

async function latestCanonicalActivity({ github, context, arcNumber, trackingIssue, arcPath, adoptionPath }) {
  const timestamps = [];
  if (trackingIssue?.updated_at) {
    timestamps.push(new Date(trackingIssue.updated_at));
  }

  const prSearch = await github.rest.search.issuesAndPullRequests({
    q: `repo:${context.repo.owner}/${context.repo.repo} is:pr "ARC-${arcNumber.toString().padStart(4, "0")}"`,
    sort: "updated",
    order: "desc",
    per_page: 1,
  });
  const pullRequest = prSearch.data.items[0];
  if (pullRequest?.updated_at) {
    timestamps.push(new Date(pullRequest.updated_at));
  }

  const gitTimestamp = await latestGitTimestamp([arcPath, adoptionPath].filter(Boolean));
  if (gitTimestamp) {
    timestamps.push(gitTimestamp);
  }

  if (timestamps.length === 0) {
    return null;
  }
  return timestamps.sort((left, right) => right - left)[0];
}

async function latestGitTimestamp(paths) {
  try {
    const { stdout } = await execFileAsync("git", ["log", "-1", "--format=%cI", "--", ...paths]);
    const trimmed = stdout.trim();
    return trimmed ? new Date(trimmed) : null;
  } catch {
    return null;
  }
}

function formatDiagnostic(diagnostic) {
  const location = diagnostic.file ? `${diagnostic.file}: ` : "";
  return `${location}${diagnostic.rule_id} ${diagnostic.message}`;
}

function renderReport(month, authorsAction, editorAction, suggestedTransitions) {
  return `# ${ISSUE_TITLE_PREFIX} ${month}

## Action required from ARC authors

${renderSection(authorsAction)}

## Action required from ARC editor

${renderSection(editorAction)}

## Suggested status transitions

${renderSection(suggestedTransitions)}
`;
}

function renderSection(items) {
  if (items.length === 0) {
    return "- No action recorded in this category for this run.";
  }
  return items.join("\n");
}

function renderRepoStateSummary(summaryMarkdown, artifactUrl) {
  const lines = ["", "## ARC state summary", ""];
  if (artifactUrl) {
    lines.push(`- Full markdown report artifact: [arc-summary.md](${artifactUrl})`);
  } else {
    lines.push("- Full markdown report artifact: unavailable for this run.");
  }

  const overview = extractRepoStateSummaryOverview(summaryMarkdown);
  if (!overview) {
    lines.push("- ARC state summary markdown was not generated.");
    return lines.join("\n");
  }

  lines.push("- Detailed transition, adoption, and relationship tables are available in the artifact.");
  lines.push("");
  lines.push(overview);

  return lines.join("\n");
}

function extractRepoStateSummaryOverview(summaryMarkdown) {
  if (!summaryMarkdown.trim()) {
    return "";
  }

  const lines = summaryMarkdown.replace(/\r\n/g, "\n").split("\n");
  const start = lines.findIndex((line) => line.trim() === "## Validation Snapshot");
  if (start === -1) {
    return "";
  }

  const end = lines.findIndex((line, index) => index > start && line.trim() === "### Counts by Status");
  const excerpt = (end === -1 ? lines.slice(start) : lines.slice(start, end)).join("\n").trim();
  return excerpt.replace(/^## /gm, "### ");
}

function authorMentions(authorLine) {
  const mentions = [...authorLine.matchAll(/@[A-Za-z0-9_-]+/g)].map((match) => match[0]);
  if (mentions.length === 0) {
    return "";
  }
  return ` Mention: ${mentions.join(" ")}.`;
}

async function upsertMonthlyIssue(github, context, month, body) {
  const title = `${ISSUE_TITLE_PREFIX} ${month}`;
  const issues = await github.paginate(github.rest.issues.listForRepo, {
    owner: context.repo.owner,
    repo: context.repo.repo,
    state: "all",
    labels: ISSUE_LABEL,
    per_page: 100,
  });
  const existing = issues.find((issue) => issue.title === title);
  if (existing) {
    await github.rest.issues.update({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: existing.number,
      title,
      body,
    });
    return;
  }
  await github.rest.issues.create({
    owner: context.repo.owner,
    repo: context.repo.repo,
    title,
    body,
    labels: [ISSUE_LABEL],
  });
}
