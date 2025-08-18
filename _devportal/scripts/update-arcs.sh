#!/bin/bash
  # Linux
  SED_INLINE=""

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
  arc_number=$(echo "$file" | grep -oE '[0-9]+')  
  if [[ -f "$file" ]]; then
    # 1. Remove the first header (and any preceding blank lines)
    sed -i $SED_INLINE '/^---$/,/^---$/!{/^# /d;}' "$file"
    if [[ $? -ne 0 ]]; then
      echo "Failed to remove header in $file"
      continue
    fi

    # 2 Step 1: Remove leading './' from links like ./arc-0001.md or ./arc-0001.md#interface-signtxnsopts
    sed -i $SED_INLINE -E 's|\(\./arc-([0-9]{1,4})\.md(\#[a-zA-Z0-9_-]+)?\)|\(arc-\1.md\2)|g' "$file"
    if [[ $? -ne 0 ]]; then
    echo "Failed to remove leading './' in links in $file"
    continue
    fi

    # 2 Step 2: Replace all arc-XXXX.md links with /arc-standards/arc-XXXX
    sed -i $SED_INLINE -E 's|\(arc-([0-9]{1,4})\.md(\#[a-zA-Z0-9_-]+)?\)|\(/arc-standards/arc-\1\2)|g' "$file"
    if [[ $? -ne 0 ]]; then
    echo "Failed to update links to /arc-standards/ in $file"
    continue
    fi
    # 3 Step 3: Replace occurrences of ': ./arc-XXXX.md#anchor' with ': /arc-standards/arc-XXXX#anchor'
    sed -i $SED_INLINE -E 's|: \./arc-([0-9]{4})\.md#([a-zA-Z0-9_-]+)|: /arc-standards/arc-\1#\2|g' "$file"
    if [[ $? -ne 0 ]]; then
    echo "Failed to update links with ': ./arc-XXXX.md#anchor' format in $file"
    continue
    fi

    # 4 Step 4: Replace occurrences of ': #anchor' with ': /arc-standards/arc-XXXX#anchor' (based on file name)
    sed -i $SED_INLINE -E "s|: #([a-zA-Z0-9_-]+)|: /arc-standards/arc-${arc_number}#\1|g" "$file"
    if [[ $? -ne 0 ]]; then
    echo "Failed to update ': #anchor' references to use /arc-standards/arc-XXXX in $file"
    continue
    fi

    # Following commands if the file is arc-0000.md
    if [[ "$file" == "arc-0000.md" ]]; then
        # Replace [ARC-0](/arc-standards/arc-0000) with [ARC-0](./arc-0000.md)
        sed -i $SED_INLINE -E 's|\[ARC-0\]\(/arc-standards/arc-0000\)|\[ARC-0\]\(./arc-0000.md\)|g' "$file"
        if [[ $? -ne 0 ]]; then
        echo "Failed to update the specific link in $file"
        continue
        fi
    fi

    # 3. Replace links like [here](../assets/arc-0062) with [here](https://github.com/algorandfoundation/ARCs/tree/main/assets/arc-0062)
    sed -i $SED_INLINE -E 's|\(\.\./assets/arc-([0-9]{4})\)|\(https://github.com/algorandfoundation/ARCs/tree/main/assets/arc-\1\)|g' "$file"
    if [[ $? -ne 0 ]]; then
      echo "Failed to update asset links in $file"
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

    #6. Replace '../assets' with 'https://raw.githubusercontent.com/algorandfoundation/ARCs/main/assets'
    sed -i $SED_INLINE "s|\(\.\./assets\)|https://raw.githubusercontent.com/algorandfoundation/ARCs/main/assets|g" "$file"
    if [[ $? -ne 0 ]]; then
      echo "Failed to replace '../assets' in $file"
      continue
    fi


  else
    echo "No markdown files found matching pattern 'arc-*.md'"
  fi
done

echo "Headers, links, and sidebar information have been successfully modified in all markdown files."
