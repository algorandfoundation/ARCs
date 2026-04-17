import fs from "node:fs/promises";
import path from "node:path";
import { execFile } from "node:child_process";
import { promisify } from "node:util";
import { parseArcMetadataFile } from "./lib/arc-metadata.mjs";
import { TRACKING_ISSUE_LABEL, hasTemplateShape } from "./lib/arc-governance.mjs";

const execFileAsync = promisify(execFile);
const SIX_MONTHS_MS = 180 * 24 * 60 * 60 * 1000;
const STAGNANT_FOLLOW_UP_MS = 210 * 24 * 60 * 60 * 1000;
const ISSUE_TITLE_PREFIX = "Monthly ARC maintenance report";
const ISSUE_LABEL = "maintenance-report";

export async function runMaintenance({ github, context, core }) {
  const offlineReport = await loadJSON(process.env.OFFLINE_REPORT_PATH);
  const onlineReport = await loadJSON(process.env.ONLINE_REPORT_PATH);
  const summaryReport = await loadText(process.env.SUMMARY_REPORT_PATH);
  const summaryArtifactUrl = process.env.SUMMARY_ARTIFACT_URL || "";
  const repoRoot = process.cwd();
  const arcDir = path.join(repoRoot, "ARCs");
  const arcFiles = (await fs.readdir(arcDir)).filter((entry) => /^arc-\d{4}\.md$/.test(entry)).sort();
  const repoIndex = await buildRepoIndex(github, context);

  const authorsAction = [];
  const editorAction = [];
  const suggestedTransitions = [];

  for (const fileName of arcFiles) {
    const filePath = path.join("ARCs", fileName);
    const metadata = await parseArcMetadataFile(path.join(repoRoot, filePath));
    if (!metadata.arc) {
      continue;
    }
    const trackingIssue = repoIndex.trackingIssues.get(normalizeArcNumber(metadata.arc)) || null;
    if (!trackingIssue) {
      editorAction.push(`- ARC-${metadata.arc} ${metadata.title || ""}: create the missing tracking issue with the \`${TRACKING_ISSUE_LABEL}\` label.`);
    } else if (!hasTemplateShape(trackingIssue.body || "")) {
      editorAction.push(`- ARC-${metadata.arc} ${metadata.title || ""}: tracking issue #${trackingIssue.number} does not match the expected template shape.`);
    }

    const latestPrUpdateAt = repoIndex.latestPrUpdates.get(normalizeArcNumber(metadata.arc)) || null;
    let latestActivity = await latestCanonicalActivity({
      trackingIssue,
      latestPrUpdateAt,
      arcPath: filePath,
      adoptionPath: metadata.adoptionSummary,
    });
    if (!latestActivity) {
      continue;
    }

    if (shouldBackfillPullRequestActivity(metadata.status, latestActivity, latestPrUpdateAt)) {
      const historicalPrUpdateAt = await findLatestHistoricalPullRequestUpdate(github, context, metadata.arc);
      if (historicalPrUpdateAt) {
        latestActivity = maxDate(latestActivity, new Date(historicalPrUpdateAt));
      }
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
    authorsAction.push(`- ${formatAuthorDiagnostic(diagnostic)}.`);
  }

  for (const diagnostic of (onlineReport.diagnostics || []).filter((item) => item.severity !== "info")) {
    authorsAction.push(`- Advisory online finding: ${formatAuthorDiagnostic(diagnostic)}.`);
  }

  const month = new Date().toISOString().slice(0, 7);
  const reportBody = renderReport(month, authorsAction, editorAction, suggestedTransitions);
  const issueBody = renderIssueBody(reportBody, summaryReport, summaryArtifactUrl);

  await core.summary
    .addHeading(`ARC maintenance report ${month}`)
    .addRaw(reportBody, true)
    .addRaw(renderRepoStateSummary(summaryReport, summaryArtifactUrl), true)
    .write();

  if (authorsAction.length === 0 && editorAction.length === 0 && suggestedTransitions.length === 0) {
    core.info("No actionable monthly maintenance findings.");
    return;
  }

  await upsertMonthlyIssue(github, context, month, issueBody);
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

async function latestCanonicalActivity({ trackingIssue, latestPrUpdateAt, arcPath, adoptionPath }) {
  const timestamps = [];
  if (trackingIssue?.updated_at) {
    timestamps.push(new Date(trackingIssue.updated_at));
  }

  if (latestPrUpdateAt) {
    timestamps.push(new Date(latestPrUpdateAt));
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

async function buildRepoIndex(github, context) {
  const [trackingIssues, pullRequests] = await Promise.all([
    github.paginate(github.rest.issues.listForRepo, {
      owner: context.repo.owner,
      repo: context.repo.repo,
      state: "all",
      labels: TRACKING_ISSUE_LABEL,
      sort: "updated",
      direction: "desc",
      per_page: 100,
    }),
    listRelevantPullRequests(github, context),
  ]);

  return {
    trackingIssues: indexTrackingIssues(trackingIssues),
    latestPrUpdates: indexPullRequestUpdates(pullRequests),
  };
}

async function listRelevantPullRequests(github, context) {
  const cutoff = Date.now() - STAGNANT_FOLLOW_UP_MS;
  const pullRequests = [];

  for await (const response of github.paginate.iterator(github.rest.pulls.list, {
    owner: context.repo.owner,
    repo: context.repo.repo,
    state: "all",
    sort: "updated",
    direction: "desc",
    per_page: 100,
  })) {
    let reachedCutoff = false;

    for (const pullRequest of response.data) {
      if (new Date(pullRequest.updated_at).getTime() < cutoff) {
        reachedCutoff = true;
        break;
      }
      pullRequests.push(pullRequest);
    }

    if (reachedCutoff) {
      break;
    }
  }

  return pullRequests;
}

function indexTrackingIssues(issues) {
  const byArc = new Map();
  for (const issue of issues) {
    if (issue.pull_request) {
      continue;
    }
    const arcNumber = extractCanonicalArcNumberFromTrackingIssue(issue);
    if (!arcNumber) {
      continue;
    }
    const current = byArc.get(arcNumber);
    if (!current || isNewerIssue(issue, current)) {
      byArc.set(arcNumber, issue);
    }
  }
  return byArc;
}

function extractCanonicalArcNumberFromTrackingIssue(issue) {
  const titleMatches = [...extractArcNumbersFromText(issue.title || "")];
  if (titleMatches.length === 1) {
    return titleMatches[0];
  }
  return "";
}

function isNewerIssue(left, right) {
  const leftUpdatedAt = new Date(left.updated_at || 0).getTime();
  const rightUpdatedAt = new Date(right.updated_at || 0).getTime();
  if (leftUpdatedAt !== rightUpdatedAt) {
    return leftUpdatedAt > rightUpdatedAt;
  }
  return (left.number || 0) > (right.number || 0);
}

function indexPullRequestUpdates(pullRequests) {
  const byArc = new Map();
  for (const pullRequest of pullRequests) {
    const arcNumber = extractCanonicalArcNumberFromPullRequest(pullRequest);
    if (!arcNumber) {
      continue;
    }
    const current = byArc.get(arcNumber);
    if (!current || new Date(pullRequest.updated_at) > new Date(current)) {
      byArc.set(arcNumber, pullRequest.updated_at);
    }
  }
  return byArc;
}

function extractCanonicalArcNumberFromPullRequest(pullRequest) {
  const titleMatches = [...extractArcNumbersFromText(pullRequest.title || "")];
  if (titleMatches.length === 1) {
    return titleMatches[0];
  }
  if (titleMatches.length > 1) {
    return "";
  }

  const bodyMatches = [...extractArcNumbersFromText(pullRequest.body || "")];
  if (bodyMatches.length === 1) {
    return bodyMatches[0];
  }
  return "";
}

function extractArcNumbersFromText(text) {
  const arcNumbers = new Set();
  for (const match of text.matchAll(/\barc[-_](\d{4})\b/gi)) {
    arcNumbers.add(match[1]);
  }
  return arcNumbers;
}

function normalizeArcNumber(value) {
  return value.toString().padStart(4, "0");
}

function shouldBackfillPullRequestActivity(status, latestActivity, latestPrUpdateAt) {
  if (latestPrUpdateAt) {
    return false;
  }

  const age = Date.now() - latestActivity.getTime();
  if (["Draft", "Review", "Last Call", "Final"].includes(status)) {
    return age >= SIX_MONTHS_MS;
  }
  if (status === "Stagnant") {
    return age >= STAGNANT_FOLLOW_UP_MS;
  }
  return false;
}

async function findLatestHistoricalPullRequestUpdate(github, context, arcNumber) {
  const normalizedArcNumber = normalizeArcNumber(arcNumber);
  const maxPages = 5;

  for (let page = 1; page <= maxPages; page += 1) {
    const result = await github.rest.search.issuesAndPullRequests({
      q: `repo:${context.repo.owner}/${context.repo.repo} is:pr "ARC-${normalizedArcNumber}"`,
      sort: "updated",
      order: "desc",
      per_page: 20,
      page,
    });
    for (const item of result.data.items) {
      if (extractCanonicalArcNumberFromPullRequest(item) === normalizedArcNumber) {
        return item.updated_at || "";
      }
    }

    if (result.data.items.length < 20) {
      break;
    }
  }
  return "";
}

function maxDate(left, right) {
  return left > right ? left : right;
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

function formatAuthorDiagnostic(diagnostic) {
  const arcNumbers = extractArcNumbersFromDiagnostic(diagnostic);
  const detail = formatDiagnostic(diagnostic);
  if (arcNumbers.length === 0) {
    return detail;
  }
  return `${arcNumbers.map((arcNumber) => `ARC-${arcNumber}`).join(", ")}: ${detail}`;
}

function extractArcNumbersFromDiagnostic(diagnostic) {
  const arcNumbers = new Set();
  const candidates = [diagnostic.file, diagnostic.message, diagnostic.hint];

  for (const candidate of candidates) {
    if (!candidate) {
      continue;
    }
    for (const arcNumber of extractArcNumbersFromText(candidate)) {
      arcNumbers.add(arcNumber);
    }
  }

  return [...arcNumbers].sort();
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

function renderIssueBody(reportBody, summaryMarkdown, artifactUrl) {
  return `${reportBody}
${renderRepoStateSummary(summaryMarkdown, artifactUrl)}
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
