#!/usr/bin/env python3

import os
import sys
import argparse
import requests
import json
import subprocess
from typing import List, Dict, Any, Optional

def get_github_token() -> str:
    """
    Retrieve a GitHub token using multiple methods.
    
    1. Check GITHUB_TOKEN environment variable
    2. Try to get token using GitHub CLI
    3. Exit with error if all methods fail
    
    Returns:
        GitHub token as string
    """
    # Method 1: Environment variable
    token = os.environ.get('GITHUB_TOKEN')
    if token:
        return token
    
    # Method 2: GitHub CLI
    try:
        result = subprocess.run(
            ["gh", "auth", "token"], 
            capture_output=True, 
            text=True, 
            check=False
        )
        if result.returncode == 0 and result.stdout.strip():
            return result.stdout.strip()
    except FileNotFoundError:
        # GitHub CLI not installed
        pass
    
    # All methods failed
    print("Error: Unable to retrieve GitHub token.", file=sys.stderr)
    print("Please either:", file=sys.stderr)
    print("  1. Set the GITHUB_TOKEN environment variable, or", file=sys.stderr)
    print("  2. Install and authenticate with GitHub CLI (gh)", file=sys.stderr)
    sys.exit(1)

def list_versions(namespace: str, package_name: str) -> List[Dict[str, Any]]:
    """
    List all versions of a package in the GitHub Container Registry.
    
    Args:
        namespace: The namespace in form of 'user/<username>' or 'org/<orgname>'
        package_name: The name of the package
    
    Returns:
        List of package versions with metadata
    """
    # Get authentication token
    token = get_github_token()
    
    # Construct API URL
    api_url = f"https://api.github.com/{namespace}/packages/container/{package_name}/versions?per_page=100"
    
    # Set up headers with authentication
    headers = {
        "Accept": "application/vnd.github+json",
        "Authorization": f"Bearer {token}",
        "X-GitHub-Api-Version": "2022-11-28"
    }
    
    try:
        response = requests.get(api_url, headers=headers)
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as e:
        print(f"Error accessing GitHub API: {e}", file=sys.stderr)
        sys.exit(1)

def main():
    parser = argparse.ArgumentParser(description="GitHub Container Registry (GHCR) CLI Tool")
    subparsers = parser.add_subparsers(dest="command", help="Commands")
    
    # List versions command
    list_parser = subparsers.add_parser("list-versions", help="List package versions")
    list_parser.add_argument("--namespace", required=True, help="Namespace in form 'user/<username>' or 'org/<orgname>'")
    list_parser.add_argument("--package", required=True, help="Package name")
    
    args = parser.parse_args()
    
    # Check if no command is provided
    if args.command is None:
        parser.print_help()
        sys.exit(1)
    
    if args.command == "list-versions":
        versions = list_versions(args.namespace, args.package)
        print(f"Versions for {args.namespace}/{args.package}:")
        for version in versions:
            print(f"  - ID: {version.get('id')}")
            print(f"    Name: {version.get('name')}")
            print(f"    Created: {version.get('created_at')}")
            print(f"    Updated: {version.get('updated_at')}")
            print(f"    Tags: {', '.join(version.get('metadata', {}).get('container', {}).get('tags', []))}")
            print("")

if __name__ == "__main__":
    main()
