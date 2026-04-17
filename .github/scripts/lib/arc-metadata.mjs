import fs from "node:fs/promises";

export async function parseArcMetadataFile(filePath) {
  const content = await fs.readFile(filePath, "utf8");
  return parseArcMetadataText(content);
}

export function parseArcMetadataText(content) {
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
