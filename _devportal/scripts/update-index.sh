#!/bin/bash

# Exit if any command fails
set -e

# Define directories and files
SRC_DIR="_devportal/content"
TEMPLATE_FILE="_devportal/scripts/index_template.md"
OUTPUT_FILE="_devportal/content/index.md"

# Ensure the source directory and template file exist
if [[ ! -d "$SRC_DIR" ]]; then
  echo "Source directory not found: $SRC_DIR"
  exit 1
fi

if [[ ! -f "$TEMPLATE_FILE" ]]; then
  echo "Template file not found: $TEMPLATE_FILE"
  exit 1
fi

# Helper function to extract a field from a markdown file
extract_field() {
  local file="$1"
  local field="$2"
  grep "^$field: " "$file" | sed "s/$field: //"
}

# Initialize arrays to hold ARCs grouped by status
declare -a living_arcs final_arcs last_call_arcs withdrawn_arcs draft_arcs stagnant_arcs review_arcs

# Loop through all markdown files in the source directory
for file in "$SRC_DIR"/arc-*.md; do
  # Extract metadata from the file
  title=$(extract_field "$file" "title")
  arc=$(extract_field "$file" "arc")
  description=$(extract_field "$file" "description")
  status=$(extract_field "$file" "status")

  # Format the ARC number to be 4 digits only for the href
  arc_href=$(printf "%04d" "$arc")

  # Prepare the formatted output for each ARC in an HTML row, adding / to hrefs and keeping ARC number as-is elsewhere
  arc_output="<tr>
    <td><a href=\"/arc-standards/arcs/arc-$arc_href/\" style='display: block; text-decoration: none; color: inherit;'>$arc</a></td>
    <td><a href=\"/arc-standards/arcs/arc-$arc_href/\" style='display: block; text-decoration: none; color: inherit;'>$title</a></td>
    <td><a href=\"/arc-standards/arcs/arc-$arc_href/\" style='display: block; text-decoration: none; color: inherit;'>$description</a></td>
  </tr>"

  # Group the ARCs by status
  case $status in
    "Living")
      living_arcs+=("$arc_output")
      ;;
    "Final")
      final_arcs+=("$arc_output")
      ;;
    "Last Call")
      last_call_arcs+=("$arc_output")
      ;;
    "Withdrawn")
      withdrawn_arcs+=("$arc_output")
      ;;
    "Draft")
      draft_arcs+=("$arc_output")
      ;;
      "Stagnant")
        stagnant_arcs+=("$arc_output")
        ;;
    "Deprecated")
        deprecated_arcs+=("$arc_output")
        ;;
    "Review")
      review_arcs+=("$arc_output")
      ;;
    *)
      echo "Unknown status: $status in $file"
      ;;
  esac
done

# Function to generate HTML table for a given set of ARCs
generate_arcs_table() {
  local title="$1"
  shift
  local arcs=("$@")

  if [ ${#arcs[@]} -eq 0 ]; then
    return # Skip empty sections
  fi

  echo "<section>"
  echo "  <h2>$title</h2>"
  echo "  <table>"
  echo "    <thead>"
  echo "      <tr>"
  echo "        <th>Number</th>"
  echo "        <th>Title</th>"
  echo "        <th>Description</th>"
  echo "      </tr>"
  echo "    </thead>"
  echo "    <tbody>"

  for arc_row in "${arcs[@]}"; do
    echo "      $arc_row"
  done

  echo "    </tbody>"
  echo "  </table>"
  echo "</section>"
}

# Step 1: Copy the template into the output file, skipping the <ArcsList> part
sed '/<ArcsList>/,$d' "$TEMPLATE_FILE" > "$OUTPUT_FILE"

# Step 2: Generate and append the HTML tables for each group of ARCs based on their status
{
  generate_arcs_table "Living Arcs" "${living_arcs[@]}"
  generate_arcs_table "Final Arcs" "${final_arcs[@]}"
  generate_arcs_table "Last Call Arcs" "${last_call_arcs[@]}"
  generate_arcs_table "Withdrawn Arcs" "${withdrawn_arcs[@]}"
  generate_arcs_table "Deprecated Arcs" "${deprecated_arcs[@]}"
  generate_arcs_table "Draft Arcs" "${draft_arcs[@]}"
  generate_arcs_table "Stagnant Arcs" "${stagnant_arcs[@]}"
  generate_arcs_table "Review Arcs" "${review_arcs[@]}"
} >> "$OUTPUT_FILE"

# Step 3: Append the rest of the template after <ArcsList>
sed -n '/<ArcsList>/,$p' "$TEMPLATE_FILE" | sed '1d' >> "$OUTPUT_FILE"

echo "Generated ARC index in $OUTPUT_FILE"
