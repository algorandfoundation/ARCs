import fs from "node:fs/promises";
import { execFile } from "node:child_process";
import { promisify } from "node:util";

const execFileAsync = promisify(execFile);
const RULE_ID = "LYCHEE";
const TITLE = "External link validation failed";
const HINT = "Review the external URLs reported by lychee.";

async function main() {
  const options = parseArgs(process.argv.slice(2));
  const command = buildCommand(options);
  const result = await runPreCommit(command);
  const diagnostics = buildDiagnostics(result.output, result.exitCode);
  const report = {
    command: "pre-commit lychee-json",
    diagnostics,
    summary: summarize(diagnostics),
    exit_code: result.exitCode,
  };
  await fs.writeFile(options.reportPath, JSON.stringify(report, null, 2) + "\n", "utf8");
  process.exit(result.exitCode);
}

function parseArgs(args) {
  const options = {
    allFiles: false,
    fromRef: "",
    toRef: "",
    reportPath: "",
  };
  for (let index = 0; index < args.length; index++) {
    switch (args[index]) {
      case "--all-files":
        options.allFiles = true;
        break;
      case "--from-ref":
        options.fromRef = args[index + 1] || "";
        index++;
        break;
      case "--to-ref":
        options.toRef = args[index + 1] || "";
        index++;
        break;
      case "--report":
        options.reportPath = args[index + 1] || "";
        index++;
        break;
      default:
        throw new Error(`unsupported argument ${args[index]}`);
    }
  }
  if (!options.reportPath) {
    throw new Error("missing --report");
  }
  if (!options.allFiles && (!options.fromRef || !options.toRef)) {
    throw new Error("use --all-files or provide both --from-ref and --to-ref");
  }
  return options;
}

function buildCommand(options) {
  const args = ["run", "lychee-json", "--hook-stage", "manual", "--color", "never"];
  if (options.allFiles) {
    args.push("--all-files");
  } else {
    args.push("--from-ref", options.fromRef, "--to-ref", options.toRef);
  }
  return args;
}

async function runPreCommit(args) {
  try {
    const { stdout, stderr } = await execFileAsync("pre-commit", args, {
      maxBuffer: 32 * 1024 * 1024,
    });
    return {
      exitCode: 0,
      output: joinOutput(stdout, stderr),
    };
  } catch (error) {
    return {
      exitCode: typeof error.code === "number" ? error.code : 1,
      output: joinOutput(error.stdout || "", error.stderr || ""),
    };
  }
}

function joinOutput(stdout, stderr) {
  return [stdout, stderr].filter(Boolean).join("\n");
}

function buildDiagnostics(output, exitCode) {
  if (exitCode === 0) {
    return [];
  }
  const structuredDiagnostics = extractStructuredDiagnostics(output);
  if (structuredDiagnostics.length > 0) {
    return structuredDiagnostics;
  }
  const diagnostics = [];
  for (const url of extractURLs(output)) {
    diagnostics.push({
      rule_id: RULE_ID,
      severity: "warning",
      title: TITLE,
      message: `external link failed: ${url}`,
      hint: HINT,
      origin: "pre-commit",
    });
  }
  if (diagnostics.length > 0) {
    return diagnostics;
  }
  return [
    {
      rule_id: RULE_ID,
      severity: "warning",
      title: TITLE,
      message: "lychee reported external link failures",
      hint: HINT,
      origin: "pre-commit",
    },
  ];
}

function extractStructuredDiagnostics(output) {
  const diagnostics = [];
  for (const payload of extractJSONPayloads(output)) {
    const created = diagnosticsFromLycheeJSON(payload);
    if (created.length > 0) {
      diagnostics.push(...created);
    }
  }
  return dedupeDiagnostics(diagnostics);
}

function extractJSONPayloads(output) {
  const payloads = [];
  const lines = output.split(/\r?\n/);
  for (const line of lines) {
    const trimmed = line.trim();
    if (!(trimmed.startsWith("{") || trimmed.startsWith("["))) {
      continue;
    }
    try {
      payloads.push(JSON.parse(trimmed));
    } catch {
      // Ignore non-JSON lines and fall back to URL extraction below.
    }
  }
  return payloads;
}

function diagnosticsFromLycheeJSON(payload) {
  const diagnostics = [];
  visitLycheeFailures(payload, ({ file, url }) => {
    diagnostics.push({
      rule_id: RULE_ID,
      severity: "warning",
      title: TITLE,
      message: `external link failed: ${url}`,
      hint: HINT,
      origin: "pre-commit",
      file,
    });
  });
  return diagnostics;
}

function visitLycheeFailures(value, visitor, currentFile = "") {
  if (!value || typeof value !== "object") {
    return;
  }

  if (Array.isArray(value)) {
    for (const entry of value) {
      visitLycheeFailures(entry, visitor, currentFile);
    }
    return;
  }

  const nextFile = pickSourceFile(value) || currentFile;
  const url = pickFailedURL(value);
  if (url) {
    visitor({ file: nextFile, url });
  }

  for (const nested of Object.values(value)) {
    visitLycheeFailures(nested, visitor, nextFile);
  }
}

function pickSourceFile(value) {
  for (const key of ["input", "file", "filename", "path", "source"]) {
    const candidate = value[key];
    if (typeof candidate === "string" && /\.[A-Za-z0-9]+$/.test(candidate)) {
      return candidate.replace(/^\.\//, "");
    }
  }
  return "";
}

function pickFailedURL(value) {
  const status = typeof value.status === "string" ? value.status.toLowerCase() : "";
  const urlKeys = ["uri", "url", "link"];
  const url = urlKeys
    .map((key) => value[key])
    .find((candidate) => typeof candidate === "string" && candidate.startsWith("http"));

  if (status && status !== "fail" && status !== "failed" && status !== "error") {
    return "";
  }
  return url || "";
}

function dedupeDiagnostics(diagnostics) {
  const byKey = new Map();
  for (const diagnostic of diagnostics) {
    const key = JSON.stringify([diagnostic.file || "", diagnostic.message]);
    if (!byKey.has(key)) {
      byKey.set(key, diagnostic);
    }
  }
  return [...byKey.values()];
}

function extractURLs(output) {
  const matches = output.match(/https?:\/\/[^\s"'`<>]+/g) || [];
  return [...new Set(matches.map((match) => match.replace(/[),.;]+$/, "")))].sort();
}

function summarize(diagnostics) {
  const summary = { errors: 0, warnings: 0, info: 0 };
  for (const diagnostic of diagnostics) {
    if (diagnostic.severity === "error") {
      summary.errors += 1;
    } else if (diagnostic.severity === "warning") {
      summary.warnings += 1;
    } else {
      summary.info += 1;
    }
  }
  return summary;
}

await main();
