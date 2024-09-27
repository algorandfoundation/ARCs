#!/bin/bash

# Detect the OS and set the appropriate sed inline flag
if [[ "$OSTYPE" == "darwin"* ]]; then
  # macOS
  SED_INLINE="''"
else
  # Linux
  SED_INLINE=""
fi

# Define source and destination directories
SRC_DIR="ARCs"
DEST_DIR="_devportal/content"

# Ensure both directories exist
if [[ ! -d "$SRC_DIR" ]]; then
  echo "Source directory not found: $SRC_DIR"
  exit 1
fi

if [[ ! -d "$DEST_DIR" ]]; then
  echo "Destination directory not found: $DEST_DIR"
  mkdir -p "$DEST_DIR"
  if [[ $? -ne 0 ]]; then
    echo "Failed to create destination directory: $DEST_DIR"
    exit 1
  fi
fi

# Step 1: Remove existing markdown files in the destination folder
echo "Removing old markdown files from $DEST_DIR..."
rm -f "$DEST_DIR"/arc*
if [[ $? -ne 0 ]]; then
  echo "Failed to remove old markdown files from $DEST_DIR"
  exit 1
fi

# Step 2: Copy markdown files from ARCs to the devportal content folder
echo "Copying markdown files from $SRC_DIR to $DEST_DIR..."
cp -r "$SRC_DIR"/* "$DEST_DIR"
if [[ $? -ne 0 ]]; then
  echo "Failed to copy markdown files from $SRC_DIR to $DEST_DIR"
  exit 1
fi

# Step 3: Modify headers and links in markdown files in the destination directory
echo "Modifying headers and links in markdown files..."

cd "$DEST_DIR" || { echo "Directory not found: $DEST_DIR"; exit 1; }

for file in arc-*.md; do
  if [[ -f "$file" ]]; then
    # # 1. Remove the first header (and any preceding blank lines)
    sed -i $SED_INLINE '/^---$/,/^---$/!{/^# /d;}' "$file"
    if [[ $? -ne 0 ]]; then
      echo "Failed to remove header in $file"
      continue
    fi

    # 2. Replace links like [ARC-1](./arc-0001.md) with [ARC-1](../arc-0001)
    sed -i $SED_INLINE -E 's|\(\./arc-([0-9]+)\.md\)|(\.\./arc-\1)|g' "$file"
    if [[ $? -ne 0 ]]; then
      echo "Failed to update links in $file"
      continue
    fi

    # 3. Handle anchors like [ARC-1](./arc-0001.md#interface-signtxnsopts) -> [ARC-1](../arc-0001#interface-signtxnsopts)
    sed -i $SED_INLINE -E 's|\(\./arc-([0-9]+)\.md(\#[a-zA-Z0-9-]+)?\)|(\.\./arc-\1\2)|g' "$file"
    if [[ $? -ne 0 ]]; then
      echo "Failed to update anchored links in $file"
      continue
    fi

    # 4. Ensure exactly one blank line between the end of the front matter and the Abstract section
    sed -i $SED_INLINE -E '/^---$/,/^## Abstract$/ { /^\s*$/d; }' "$file"
    if [[ $? -ne 0 ]]; then
      echo "Failed to remove extra blank lines in $file"
      continue
    fi

    # Step 5: Modify the markdown file to add sidebar, label, and badge based on the status
    status=$(grep '^status: ' "$file" | sed 's/status: //')
    filename=$(basename -- "$file")
    arc_label="${filename%.*}"

    case $status in
      "Final")
        variant="success"
        ;;
      "Draft")
        variant="caution"
        ;;
      "Last Call")
        variant="note"
        ;;
      "Withdrawn")
        variant="danger"
        ;;
      "Deprecated")
        variant="danger"
        ;;
      *)
        variant="tip"
        ;;
    esac

    # Add the sidebar information right after the 'status:' line
    sed -i $SED_INLINE "/^status: /a\\
sidebar:\\
  label: $arc_label\\
  badge:\\
    text: $status\\
    variant: $variant" "$file"
    if [[ $? -ne 0 ]]; then
      echo "Failed to add sidebar information in $file"
      continue
    fi

  else
    echo "No markdown files found matching pattern 'arc-*.md'"
  fi
done

echo "Headers, links, and sidebar information have been successfully modified in all markdown files."