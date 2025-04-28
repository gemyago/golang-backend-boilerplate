#!/usr/bin/env bash

# This script tags existing remote Docker images using the crane tool.
# It reads a list of base image names (e.g., ghcr.io/namespace/repo-app)
# from a specified file. It determines the target tags based on the
# target commit SHA and git reference by calling resolve-docker-tags.sh.
# Then, it uses 'crane tag' to apply these target tags to the images
# originally tagged with the source commit SHA.
# NOTE: Both source and target commit SHAs are automatically shortened
#       to the first 7 characters internally for tagging purposes.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# --- Configuration ---
# Assumed relative paths from the repository root
readonly CRANE_PATH="${SCRIPT_DIR}/../bin/crane"
readonly RESOLVE_TAGS_SCRIPT_PATH="${SCRIPT_DIR}/resolve-docker-tags.sh"
readonly BUILD_CFG_PATH="${SCRIPT_DIR}/../build.cfg"
# Default location for the file listing remote image base names
DEFAULT_REMOTE_IMAGES_FILE="${SCRIPT_DIR}/../docker/.remote-images"

# --- Script Functions ---

# Function to print usage information
usage() {
  echo "Usage: $0 --source-commit-sha <sha> --target-commit-sha <sha> --git-ref <ref> [options]"
  echo ""
  echo "Required Arguments:"
  echo "  --source-commit-sha <sha>   The commit SHA the images are currently tagged with."
  echo "  --target-commit-sha <sha>   The commit SHA to use for calculating the target tags."
  echo "  --git-ref <ref>               The Git reference (branch or tag) corresponding to the target commit."
  echo ""
  echo "Options:"
  echo "  --remote-images-file <path> Path to the file containing base image names (default: ${DEFAULT_REMOTE_IMAGES_FILE})"
  echo "  --noop                        Print the crane commands instead of executing them."
  echo "  -h, --help                    Show this help message"
}

# Function to read a value from the build.cfg file
# Usage: read_config <key_name>
read_config() {
    local key="$1"
    local value
    value=$(grep -m 1 "^${key}\s*=\s*" "${BUILD_CFG_PATH}" | cut -d'=' -f2 | sed 's/^[[:space:]]*//;s/[[:space:]]*$$//')
    if [[ -z "$value" ]]; then
        echo "Error: Key '${key}' not found or empty in ${BUILD_CFG_PATH}" >&2
        return 1
    fi
    echo "$value"
}

# --- Argument Parsing ---
REMOTE_IMAGES_FILE="${DEFAULT_REMOTE_IMAGES_FILE}"
SOURCE_COMMIT_SHA=""
TARGET_COMMIT_SHA=""
GIT_REF=""
NOOP=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --remote-images-file)
      REMOTE_IMAGES_FILE="$2"
      shift 2
      ;;
    --source-commit-sha)
      SOURCE_COMMIT_SHA="$2"
      shift 2
      ;;
    --target-commit-sha)
      TARGET_COMMIT_SHA="$2"
      shift 2
      ;;
    --git-ref)
      GIT_REF="$2"
      shift 2
      ;;
    --noop)
      NOOP=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage
      exit 1
      ;;
  esac
done

# --- Input Validation ---
if [[ -z "$SOURCE_COMMIT_SHA" ]]; then
  echo "Error: --source-commit-sha is required" >&2
  usage
  exit 1
fi

if [[ -z "$TARGET_COMMIT_SHA" ]]; then
  echo "Error: --target-commit-sha is required" >&2
  usage
  exit 1
fi

if [[ -z "$GIT_REF" ]]; then
  echo "Error: --git-ref is required" >&2
  usage
  exit 1
fi

# --- Derive Short SHAs ---
SHORT_SOURCE_COMMIT_SHA="${SOURCE_COMMIT_SHA:0:7}"
SHORT_TARGET_COMMIT_SHA="${TARGET_COMMIT_SHA:0:7}"

if [[ ! -f "$REMOTE_IMAGES_FILE" ]]; then
    echo "Error: Remote images file not found: ${REMOTE_IMAGES_FILE}" >&2
    exit 1
fi

if [[ ! -x "$CRANE_PATH" ]] && [[ "$NOOP" == "false" ]]; then # Only check if crane exists if not noop
    echo "Error: crane executable not found or not executable: ${CRANE_PATH}" >&2
    echo "Please ensure crane is installed, e.g., by running 'make build/bin/crane'" >&2
    exit 1
fi

if [[ ! -x "$RESOLVE_TAGS_SCRIPT_PATH" ]]; then
    echo "Error: resolve-docker-tags.sh script not found or not executable: ${RESOLVE_TAGS_SCRIPT_PATH}" >&2
    exit 1
fi

if [[ ! -f "$BUILD_CFG_PATH" ]]; then
    echo "Error: Build config file not found: ${BUILD_CFG_PATH}" >&2
    exit 1
fi

# --- Main Logic ---

echo "--- Reading Configuration ---"
STABLE_BRANCHES=$(read_config "stable_branches")
if [[ $? -ne 0 ]]; then
    exit 1
fi
echo "Using stable branches: ${STABLE_BRANCHES}"

echo ""
echo "--- Determining Target Tags ---"
echo "Calculating tags for Ref: ${GIT_REF}, Commit: ${SHORT_TARGET_COMMIT_SHA}"
TARGET_TAGS=$("${RESOLVE_TAGS_SCRIPT_PATH}" \
    --commit-sha "${SHORT_TARGET_COMMIT_SHA}" \
    --git-ref "${GIT_REF}" \
    --stable-branches "${STABLE_BRANCHES}")
echo "Target tags to apply: ${TARGET_TAGS}"

echo ""
echo "--- Tagging Remote Images ---"
if [[ "$NOOP" == "true" ]]; then
    echo "*** NOOP mode enabled: Printing commands instead of executing them ***"
fi
source_image_tag="git-commit-${SHORT_SOURCE_COMMIT_SHA}"
echo "Source image tag: ${source_image_tag}"
echo "Reading base images from: ${REMOTE_IMAGES_FILE}"

tagging_errors=0
while IFS= read -r base_image || [[ -n "$base_image" ]]; do
    source_image="${base_image}:${source_image_tag}"
    if [[ -z "$base_image" ]]; then
        continue # Skip empty lines
    fi

    echo "Processing base image: ${base_image}"

    for target_tag in $TARGET_TAGS; do
        tag_command="'${CRANE_PATH}' tag '${source_image}' '${target_tag}'"

        if [[ "$NOOP" == "true" ]]; then
            echo "  [NOOP] Would run: ${tag_command}"
        else
            echo "  Attempting tag: ${source_image} -> ${target_tag}"
            if eval "${tag_command}"; then # Use eval to correctly handle quotes in paths/tags
                echo "    Success."
            else
                echo "    Error: Failed to apply tag ${target_tag} to ${base_image} (Source SHA: ${SHORT_SOURCE_COMMIT_SHA}). Check crane output above." >&2
                ((tagging_errors++))
            fi
        fi
    done
    echo ""

done < "${REMOTE_IMAGES_FILE}"

echo "--- Tagging Summary ---"
if [[ "$NOOP" == "true" ]]; then
    echo "NOOP mode finished. No changes were made."
    exit 0
fi

if [[ $tagging_errors -eq 0 ]]; then
    echo "All target tags applied successfully."
    exit 0
else
    echo "${tagging_errors} error(s) occurred during tagging." >&2
    exit 1
fi 