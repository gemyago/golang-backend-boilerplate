#!/usr/bin/env bash

# Docker Tag Resolution Logic:
# - For git tags: Uses the tag name
# - For non-stable branches: Uses branch name and git-commit-<commit-sha>
# - For stable branches: Uses latest-<branch-name> and git-commit-<commit-sha>
# All generated tags are sanitized to match Docker's tag format requirements.

set -euo pipefail

# Function to print usage information
usage() {
  echo "Usage: $0 [options]"
  echo "Options:"
  echo "  --stable-branches <branches>  Comma-separated list of stable branches"
  echo "  --git-ref <ref>               Git reference (branch or tag)"
  echo "  --commit-sha <sha>            Git commit SHA"
  echo "  --self-test                   Run internal tests"
  echo "  -h, --help                    Show this help message"
}

# Function to sanitize Docker tag according to the pattern /[\w][\w.-]{0,127}/
# First character must be alphanumeric, remaining can be alphanumeric, dot, or dash
sanitize_docker_tag() {
  local tag="$1"
  local sanitized=""
  
  # First character must be alphanumeric
  if [[ "${tag:0:1}" =~ [a-zA-Z0-9] ]]; then
    sanitized="${tag:0:1}"
  else
    sanitized="x-"
  fi
  
  local remaining="${tag:1}"
  local sanitized_remaining=$(echo "$remaining" | sed -E 's/[^a-zA-Z0-9.-]/-/g')
  
  sanitized="${sanitized}${sanitized_remaining}"
  echo "${sanitized:0:128}"
}

# Function to resolve Docker tags based on git reference
resolve_docker_tags() {
  local ref=$1
  local sha=$2
  local stable_branches=$3
  local tags=""

  # If ref is a tag (refs/tags/v1.0.0), return the tag without 'refs/tags/' prefix
  if [[ $ref == refs/tags/* ]]; then
    local tag_name="${ref#refs/tags/}"
    tags="$(sanitize_docker_tag "$tag_name")"
  # If ref is a branch (refs/heads/main), process accordingly
  elif [[ $ref == refs/heads/* ]]; then
    local branch_name="${ref#refs/heads/}"
    local is_stable=false
    local sanitized_branch="$(sanitize_docker_tag "$branch_name")"
    local sanitized_commit="git-commit-$(sanitize_docker_tag "$sha")"
    
    # Check if branch is in the list of stable branches
    IFS=',' read -ra STABLE_BRANCHES <<< "$stable_branches"
    for stable in "${STABLE_BRANCHES[@]}"; do
      if [[ "$stable" == "$branch_name" ]]; then
        is_stable=true
        break
      fi
    done
    
    if [[ "$is_stable" == "true" ]]; then
      # For stable branches: latest-<branch-name> and git-commit-<commit-sha>
      tags="latest-${sanitized_branch},${sanitized_commit}"
    else
      # For non-stable branches: <branch-name> and git-commit-<commit-sha>
      tags="${sanitized_branch},${sanitized_commit}"
    fi
  fi

  echo "$tags"
}

run_tests() {
  assert() {
    local actual="$1"
    local expected="$2"
    local message="$3"
    
    if [[ "$actual" == "$expected" ]]; then
      echo "PASS: $message"
      return 0
    else
      echo "FAIL: $message"
      echo "  Expected: '$expected'"
      echo "  Actual:   '$actual'"
      return 1
    fi
  }

  echo "Running self-tests..."
  local failures=0
  local test_result
  
  echo ""
  echo "Running sanitize_docker_tag tests:"
  test_result=$(sanitize_docker_tag "valid-tag")
  assert "$test_result" "valid-tag" "Simple valid tag" || ((failures++))
  
  test_result=$(sanitize_docker_tag "_invalid-first-char")
  assert "$test_result" "x-invalid-first-char" "Invalid first character" || ((failures++))
  
  test_result=$(sanitize_docker_tag "branch/with/slashes")
  assert "$test_result" "branch-with-slashes" "Replace slashes with dashes" || ((failures++))
  
  test_result=$(sanitize_docker_tag "tag@with#special&chars")
  assert "$test_result" "tag-with-special-chars" "Replace special chars with dashes" || ((failures++))
  
  test_result=$(sanitize_docker_tag "very-long-tag-$(printf '%0.s-' {1..150})")
  assert "${#test_result}" "128" "Truncate long tag to 128 chars" || ((failures++))
  
  local resolve_result
  local commit_sha="abc1234567890"

  echo ""
  echo "Running resolve_docker_tags tests:"
  resolve_result=$(resolve_docker_tags "refs/tags/v1.0.0" "$commit_sha" "main,develop")
  assert "$resolve_result" "v1.0.0" "Standard tag resolution" || ((failures++))
  
  resolve_result=$(resolve_docker_tags "refs/heads/feature/xyz" "$commit_sha" "main,develop")
  assert "$resolve_result" "feature-xyz,git-commit-$commit_sha" "Standard non-stable branch resolution with sanitization" || ((failures++))
  
  resolve_result=$(resolve_docker_tags "refs/heads/main" "$commit_sha" "main,develop")
  assert "$resolve_result" "latest-main,git-commit-$commit_sha" "Standard stable branch resolution" || ((failures++))
  
  # Test with special characters
  resolve_result=$(resolve_docker_tags "refs/tags/v1.0.0@special" "$commit_sha" "main,develop")
  assert "$resolve_result" "v1.0.0-special" "Tag with special characters" || ((failures++))
  
  resolve_result=$(resolve_docker_tags "refs/heads/feature/special@chars#here" "$commit_sha" "main,develop")
  assert "$resolve_result" "feature-special-chars-here,git-commit-$commit_sha" "Branch with special characters" || ((failures++))
  
  if [[ $failures -eq 0 ]]; then
    echo "All tests passed!"
    return 0
  else
    echo "$failures test(s) failed."
    return 1
  fi
}

STABLE_BRANCHES=""
GIT_REF=""
COMMIT_SHA=""
SELF_TEST=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --stable-branches)
      STABLE_BRANCHES="$2"
      shift 2
      ;;
    --git-ref)
      GIT_REF="$2"
      shift 2
      ;;
    --commit-sha)
      COMMIT_SHA="$2"
      shift 2
      ;;
    --self-test)
      SELF_TEST=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      usage
      exit 1
      ;;
  esac
done

if $SELF_TEST; then
  run_tests
  exit $?
fi

if [[ -z "$GIT_REF" ]]; then
  echo "Error: --git-ref is required"
  usage
  exit 1
fi

if [[ -z "$COMMIT_SHA" ]]; then
  echo "Error: --commit-sha is required"
  usage
  exit 1
fi

tags=$(resolve_docker_tags "$GIT_REF" "$COMMIT_SHA" "$STABLE_BRANCHES")
echo "$tags"