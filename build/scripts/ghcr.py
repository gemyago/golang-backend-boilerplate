#!/usr/bin/env python3

import os
import sys
import argparse
import requests
import json
from typing import List, Dict, Any, Optional

def list_versions(namespace: str, package_name: str) -> List[Dict[str, Any]]:
    """
    List all versions of a package in the GitHub Container Registry.
    
    Args:
        namespace: The namespace in form of 'user/<username>' or 'org/<orgname>'
        package_name: The name of the package
    
    Returns:
        List of package versions with metadata
    """
    # Get authentication token from environment
    token = os.environ.get('GITHUB_TOKEN')
    if not token:
        print("Error: GITHUB_TOKEN environment variable is required", file=sys.stderr)
        sys.exit(1)
    
    # Construct API URL
    api_url = f"https://api.github.com/packages/container/{namespace}/{package_name}/versions"
    
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
