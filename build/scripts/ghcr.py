#!/usr/bin/env python3

import os
import sys
import argparse
import requests
import subprocess
from typing import List, Optional, TypedDict, Callable, Protocol, Dict, Literal, Any
from datetime import datetime, timezone
import logging

class AuthenticationError(Exception):
    """Raised when authentication fails"""
    pass

class ContainerMetadata(TypedDict):
    tags: List[str]

class VersionMetadata(TypedDict):
    container: ContainerMetadata

class PackageVersion(TypedDict):
    id: int
    name: Optional[str]
    url: str
    package_html_url: str
    created_at: str
    updated_at: str
    html_url: str
    metadata: VersionMetadata

class CleanupAction(TypedDict):
    version: PackageVersion
    action: Literal["keep", "delete"]
    reason: str

class GitHubTokenProvider:
    """
    Class responsible for retrieving and caching GitHub tokens from various sources.
    """
    def __init__(self, environ=os.environ, subprocess_module=subprocess):
        """
        Initialize the token provider.
        
        Args:
            environ: Environment dictionary to use (default: os.environ)
            subprocess_module: Subprocess module to use (default: subprocess)
        """
        self._environ = environ
        self._subprocess = subprocess_module
        self._token = None
    
    def get_token(self) -> str:
        """
        Retrieve a GitHub token using multiple methods.
        
        1. Return cached token if available
        2. Check GITHUB_TOKEN environment variable
        3. Try to get token using GitHub CLI
        4. Raise AuthenticationError if all methods fail
        
        Returns:
            GitHub token as string
            
        Raises:
            AuthenticationError: If unable to retrieve a valid GitHub token
        """
        # Return cached token if we already retrieved it
        if self._token:
            return self._token
        
        # Method 1: Environment variable
        token = self._environ.get('GITHUB_TOKEN')
        if token:
            self._token = token
            return token
        
        # Method 2: GitHub CLI
        try:
            result = self._subprocess.run(
                ["gh", "auth", "token"], 
                capture_output=True, 
                text=True, 
                check=False
            )
            if result.returncode == 0 and result.stdout.strip():
                self._token = result.stdout.strip()
                return self._token
        except FileNotFoundError:
            # GitHub CLI not installed
            pass
        
        # All methods failed
        error_msg = (
            "Unable to retrieve GitHub token.\n"
            "Please either:\n"
            "  1. Set the GITHUB_TOKEN environment variable, or\n"
            "  2. Install and authenticate with GitHub CLI (gh)"
        )
        raise AuthenticationError(error_msg)
    
    def clear_token(self):
        """
        Clear the cached token, forcing a refresh on next get_token() call.
        """
        self._token = None

# Create a default token provider instance
default_token_provider = GitHubTokenProvider()

def list_versions(namespace: str, package_name: str, token_provider=default_token_provider, requests_module=requests) -> List[PackageVersion]:
    """
    List all versions of a package in the GitHub Container Registry.
    
    Args:
        namespace: The namespace in form of 'user/<username>' or 'org/<orgname>'
        package_name: The name of the package
        token_provider: Provider that returns a GitHub token (default: default_token_provider)
        requests_module: Module to use for HTTP requests (default: requests)
    
    Returns:
        List of package versions with metadata
        
    Raises:
        AuthenticationError: If authentication fails
        APIError: If API request fails
    """
    # Get authentication token
    token = token_provider.get_token()
    
    # Construct API URL
    api_url = f"https://api.github.com/{namespace}/packages/container/{package_name}/versions?per_page=100"
    
    # Set up headers with authentication
    headers = {
        "Accept": "application/vnd.github+json",
        "Authorization": f"Bearer {token}",
        "X-GitHub-Api-Version": "2022-11-28"
    }
    
    response = requests_module.get(api_url, headers=headers)
    response.raise_for_status()
    return response.json()

def find_versions_to_clean(versions: List[PackageVersion], keep_latest: int = 5) -> List[CleanupAction]:
    """
    Find package versions that should be cleaned up/removed.
    Keeps only the latest 'keep_latest' versions that have tags.
    
    Args:
        versions: List of package versions to analyze
        keep_latest: Number of most recent tagged versions to keep (default: 5)
    
    Returns:
        List of cleanup actions with version, action ("keep" or "delete"), and reason
    """
    # Separate versions into tagged and untagged
    tagged_versions = []
    untagged_versions = []
    
    for version in versions:
        tags = version.get('metadata', {}).get('container', {}).get('tags', [])
        if tags:
            tagged_versions.append(version)
        else:
            untagged_versions.append(version)
    
    # Sort tagged versions by creation date (newest first)
    sorted_tagged = sorted(
        tagged_versions, 
        key=lambda v: datetime.fromisoformat(v['created_at'].replace('Z', '+00:00')),
        reverse=True
    )
    
    # Initialize result list
    cleanup_actions = []
    
    # Process tagged versions to keep
    for i, version in enumerate(sorted_tagged):
        if i < keep_latest:
            cleanup_actions.append({
                "version": version,
                "action": "keep",
                "reason": f"Tagged version within the {keep_latest} most recent"
            })
        else:
            cleanup_actions.append({
                "version": version,
                "action": "delete",
                "reason": f"Older tagged version beyond the {keep_latest} most recent to keep"
            })
    
    # Process untagged versions (all should be deleted)
    for version in untagged_versions:
        cleanup_actions.append({
            "version": version,
            "action": "delete",
            "reason": "Untagged version"
        })
    
    return cleanup_actions

def remove_version(namespace: str, package_name: str, version_id: int, 
                   dry_run: bool = True, token_provider=default_token_provider, 
                   requests_module=requests) -> bool:
    """
    Remove a specific package version from the GitHub Container Registry.
    
    Args:
        namespace: The namespace in form of 'user/<username>' or 'org/<orgname>'
        package_name: The name of the package
        version_id: The ID of the version to remove
        dry_run: If True, only simulate removal (default: True)
        token_provider: Provider that returns a GitHub token (default: default_token_provider)
        requests_module: Module to use for HTTP requests
    
    Returns:
        True if the version was removed (or would have been in dry_run mode)
        
    Raises:
        AuthenticationError: If authentication fails
        requests.exceptions.HTTPError: If API request fails
    """
    if dry_run:
        logging.info(f"[DRY RUN] Would remove version {version_id} of {namespace}/{package_name}")
        return True
    
    # Get authentication token
    token = token_provider.get_token()
    
    # Construct API URL
    api_url = f"https://api.github.com/{namespace}/packages/container/{package_name}/versions/{version_id}"
    
    # Set up headers with authentication
    headers = {
        "Accept": "application/vnd.github+json",
        "Authorization": f"Bearer {token}",
        "X-GitHub-Api-Version": "2022-11-28"
    }
    
    # Perform deletion
    response = requests_module.delete(api_url, headers=headers)
    response.raise_for_status()
    
    logging.info(f"Removed version {version_id} of {namespace}/{package_name}")
    return True

class CleanupArgs(Protocol):
    """Type definition for cleanup command arguments."""
    namespace: str
    package: str
    keep_latest: int
    really_remove: bool

def cleanup_versions_command(args: CleanupArgs, 
                           list_versions_func: Callable[[str, str], List[PackageVersion]] = list_versions, 
                           find_versions_func: Callable[[List[PackageVersion], int], List[CleanupAction]] = find_versions_to_clean, 
                           remove_version_func: Callable[[str, str, int, bool], bool] = remove_version):
    """
    Handle the cleanup-versions command logic.
    
    Args:
        args: Command line arguments from argparse with needed attributes
        list_versions_func: Function to list versions (default: list_versions)
        find_versions_func: Function to find versions to clean (default: find_versions_to_clean)
        remove_version_func: Function to remove a version (default: remove_version)
    """
    # Get all versions
    all_versions = list_versions_func(args.namespace, args.package)
    
    # Find versions to clean
    cleanup_actions = find_versions_func(
        all_versions, 
        keep_latest=args.keep_latest
    )
    
    print(f"Found {len(cleanup_actions)} versions from {args.namespace}/{args.package}:")
    for action in cleanup_actions:
      version = action["version"]
      name_display = version['name'] if version['name'] else 'N/A'
      tags = version['metadata']['container']['tags'] if 'container' in version['metadata'] else []
      print(f"  - ID: {version['id']}, Name: {name_display}, Tags: {', '.join(tags)}")
      print(f"    Action: {action['action']}")
      print(f"    Reason: {action['reason']}")

    # Get versions to remove (those with action="delete")
    to_remove = [action["version"] for action in cleanup_actions if action["action"] == "delete"]
    
    # Show what would be removed
    if not to_remove:
        print(f"No versions to remove from {args.namespace}/{args.package}.")
        return
    
    print(f"Removing {len(to_remove)} (really_remove: {args.really_remove}) versions from {args.namespace}/{args.package}:")
    for version in to_remove:
        remove_version_func(args.namespace, args.package, version['id'], dry_run=not args.really_remove)
    print(f"Successfully removed {len(to_remove)} versions.")

def main():
    parser = argparse.ArgumentParser(description="GitHub Container Registry (GHCR) CLI Tool")
    subparsers = parser.add_subparsers(dest="command", help="Commands")
    
    # List versions command
    list_parser = subparsers.add_parser("list-versions", help="List package versions")
    list_parser.add_argument("--namespace", required=True, help="Namespace in form 'user/<username>' or 'org/<orgname>'")
    list_parser.add_argument("--package", required=True, help="Package name")
    
    # Cleanup versions command
    cleanup_parser = subparsers.add_parser("cleanup-versions", help="Clean up old package versions")
    cleanup_parser.add_argument("--namespace", required=True, help="Namespace in form 'user/<username>' or 'org/<orgname>'")
    cleanup_parser.add_argument("--package", required=True, help="Package name")
    cleanup_parser.add_argument("--keep-latest", type=int, default=5, help="Number of latest versions to keep (default: 5)")
    cleanup_parser.add_argument("--really-remove", action="store_true", help="Actually perform deletion (without this flag, dry run is performed)")
    
    args = parser.parse_args()
    
    # Configure logging
    logging.basicConfig(level=logging.INFO, format='%(levelname)s: %(message)s')
    
    # Check if no command is provided
    if args.command is None:
        parser.print_help()
        sys.exit(1)
    
    try:
        if args.command == "list-versions":
            versions = list_versions(args.namespace, args.package)
            print(f"Versions for {args.namespace}/{args.package}:")
            for version in versions:
                print(f"  - ID: {version['id']}")
                print(f"    Name: {version['name'] if version['name'] else 'N/A'}")
                print(f"    Created: {version['created_at']}")
                print(f"    Updated: {version['updated_at']}")
                tags = version['metadata']['container']['tags'] if 'container' in version['metadata'] else []
                print(f"    Tags: {', '.join(tags)}")
                print("")
        
        elif args.command == "cleanup-versions":
            cleanup_versions_command(args)
    
    except Exception as e:
        print(f"Command failed: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
