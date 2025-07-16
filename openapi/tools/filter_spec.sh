#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 2 ]; then
  echo "Usage: $0 <input_spec.yaml> <output_filtered_spec.yaml>"
  exit 1
fi

INPUT="$1"
OUTPUT="$2"

# Define components that should be ignored during comparison.
WHITELIST=("WhitelistedComponent1" "WhitelistedComponent2")

# Declare global arrays.
declare -a pending
declare -a processed_list

# Start with ExperimentExecutionAO.
pending=("ExperimentExecutionAO")
processed_list=()

# Function to check if a component is whitelisted.
is_whitelisted() {
  local comp="$1"
  for w in "${WHITELIST[@]}"; do
    if [[ "$comp" == "$w" ]]; then
      return 0
    fi
  done
  return 1
}

# Function to check if a component has been processed.
is_processed() {
  local comp="$1"
  for p in "${processed_list[@]-}"; do
    if [[ "$p" == "$comp" ]]; then
      return 0
    fi
  done
  return 1
}

# Create a temporary directory to store component definitions.
tmp_dir=$(mktemp -d)
echo "Using temporary directory: $tmp_dir"

while [ ${#pending[@]} -gt 0 ]; do
  # Get the first component from the pending list.
  comp="${pending[0]}"
  pending=("${pending[@]:1}")

  # Skip if already processed.
  if is_processed "$comp"; then
    continue
  fi

  # Skip if component is whitelisted.
  if is_whitelisted "$comp"; then
    echo "Skipping whitelisted component: $comp"
    processed_list+=("$comp")
    continue
  fi

  # Extract the component definition using yq.
  comp_def=$(yq eval ".components.schemas.\"$comp\"" "$INPUT")
  if [ "$comp_def" == "null" ]; then
    echo "Component $comp not found in $INPUT" >&2
    exit 1
  fi

  # Write the component definition to a file.
  echo "$comp_def" > "$tmp_dir/$comp.yaml"
  processed_list+=("$comp")

  # Recursively find all $ref values in the component.
  refs=$(echo "$comp_def" | yq eval '.. | select(has("$ref")) | ."$ref"' -)
  for ref in $refs; do
    if [[ "$ref" =~ ^\#\/components\/schemas\/(.+)$ ]]; then
      ref_comp="${BASH_REMATCH[1]}"
      if ! is_processed "$ref_comp"; then
        pending+=("$ref_comp")
      fi
    fi
  done
done

# Build the filtered OpenAPI spec.
{
  echo "openapi: \"3.0.0\""
  echo "info:"
  echo "  title: Filtered Spec"
  echo "  version: \"1.0.0\""
  echo "components:"
  echo "  schemas:"
  # List components in sorted order for consistency.
  for comp_file in $(ls "$tmp_dir" | sort); do
    comp_name="${comp_file%.yaml}"
    echo "    $comp_name:"
    # Indent the component definition.
    sed 's/^/      /' "$tmp_dir/$comp_file"
  done
} > "$OUTPUT"

echo "Filtered spec written to $OUTPUT"
rm -rf "$tmp_dir"