#!/usr/bin/env bash

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

# Function to resolve Docker tags based on git reference
resolve_docker_tags() {
  local ref=$1
  local sha=$2
  local stable_branches=$3
  local tags=""

  # If ref is a tag (refs/tags/v1.0.0), return the tag without 'refs/tags/' prefix
  if [[ $ref == refs/tags/* ]]; then
    tags="${ref#refs/tags/}"
  # If ref is a branch (refs/heads/main), process accordingly
  elif [[ $ref == refs/heads/* ]]; then
    local branch_name="${ref#refs/heads/}"
    local is_stable=false
    
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
      tags="latest-${branch_name},git-commit-${sha}"
    else
      # For non-stable branches: <branch-name> and git-commit-<commit-sha>
      tags="${branch_name},git-commit-${sha}"
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
  
  local resolve_result
  local commit_sha="abc1234567890"

  # Test case 1: Tag reference
  resolve_result=$(resolve_docker_tags "refs/tags/v1.0.0" "$commit_sha" "main,develop")
  assert "$resolve_result" "v1.0.0" "Tag resolution" || ((failures++))
  
  # Test case 2: Branch reference (not stable)
  resolve_result=$(resolve_docker_tags "refs/heads/feature/xyz" "$commit_sha" "main,develop")
  assert "$resolve_result" "feature/xyz,git-commit-$commit_sha" "Non-stable branch resolution" || ((failures++))
  
  # Test case 3: Branch reference (stable)
  resolve_result=$(resolve_docker_tags "refs/heads/main" "$commit_sha" "main,develop")
  assert "$resolve_result" "latest-main,git-commit-$commit_sha" "Stable branch resolution (main)" || ((failures++))
  
  resolve_result=$(resolve_docker_tags "refs/heads/develop" "$commit_sha" "main,develop")
  assert "$resolve_result" "latest-develop,git-commit-$commit_sha" "Stable branch resolution (develop)" || ((failures++))
  
  if [[ $failures -eq 0 ]]; then
    echo "All tests passed!"
    return 0
  else
    echo "$failures test(s) failed."
    return 1
  fi
}

# Main script starts here
STABLE_BRANCHES=""
GIT_REF=""
COMMIT_SHA=""
SELF_TEST=false

# Parse command line arguments using getopts
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

# Run tests if in self-test mode
if $SELF_TEST; then
  run_tests
  exit $?
fi

# Validate required arguments are provided
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

# If not in self-test mode, resolve and output Docker tags
tags=$(resolve_docker_tags "$GIT_REF" "$COMMIT_SHA" "$STABLE_BRANCHES")
echo "$tags"