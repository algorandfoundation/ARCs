#!/bin/bash

# Exit the script if any command fails
set -e

# Define directories and files
SRC_DIR="_devportal/content"
TEMPLATE_FILE="_devportal/scripts/guidelines_template.md"
OUTPUT_FILE="_devportal/content/guideline.md"

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

# Helper function to extract abstract (assumes abstract follows the 'Abstract' header)
extract_abstract() {
  local file="$1"
  # Extract the content between "## Abstract" and the next "##", then replace "###" with "####"
  sed -n '/^## Abstract/,/^## [^#]/p' "$file" | sed '1d;$d' | sed '/^\s*$/d' | sed 's/^### /#### /'
}

# Initialize arrays to hold ARCs grouped by sub-category
declare -a general_arcs asa_arcs application_arcs explorer_arcs wallet_arcs

# Loop through all markdown files in the source directory
for file in "$SRC_DIR"/arc-*.md; do
  # Extract metadata from the file
  title=$(extract_field "$file" "title")
  arc=$(extract_field "$file" "arc")
  sub_category=$(extract_field "$file" "sub-category")
  abstract=$(extract_abstract "$file")

  # Prepare the formatted output for each ARC
  if [[ -n $abstract ]]; then
    arc_output="### ARC $arc - $title\n\n$abstract\n"
  else
    arc_output="### ARC $arc - $title\n\nNo abstract available.\n"
  fi

  # Group the ARCs by sub-category
  case $sub_category in
    "General")
      general_arcs+=("$arc_output")
      ;;
    "Asa")
      asa_arcs+=("$arc_output")
      ;;
    "Application")
      application_arcs+=("$arc_output")
      ;;
    "Explorer")
      explorer_arcs+=("$arc_output")
      ;;
    "Wallet")
      wallet_arcs+=("$arc_output")
      ;;
    *)
      echo "No sub-category: $sub_category in $file"
      ;;
  esac
done

# Function to append ARCs to the output file based on the category
append_arcs() {
  local category_arcs=("$@")
  for arc_content in "${category_arcs[@]}"; do
    echo -e "$arc_content" >> "$OUTPUT_FILE"
  done
}

# Step 1: Copy the template into the output file, skipping the <ArcsGuideList> part
sed '/<ArcsGuideList>/q' "$TEMPLATE_FILE" > "$OUTPUT_FILE"

# Step 2: Append ARCs to the output file, grouped by sub-category
echo -e "## General ARCs\n" >> "$OUTPUT_FILE"
append_arcs "${general_arcs[@]}"

echo -e "## Asa ARCs\n" >> "$OUTPUT_FILE"
append_arcs "${asa_arcs[@]}"

echo -e "## Application ARCs\n" >> "$OUTPUT_FILE"
append_arcs "${application_arcs[@]}"

echo -e "## Explorer ARCs\n" >> "$OUTPUT_FILE"
append_arcs "${explorer_arcs[@]}"

echo -e "## Wallet ARCs\n" >> "$OUTPUT_FILE"
append_arcs "${wallet_arcs[@]}"

echo "Generated ARC guidelines in $OUTPUT_FILE"
