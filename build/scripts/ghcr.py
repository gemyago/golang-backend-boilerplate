#!/usr/bin/env python3

import os
import sys
import argparse
import requests
import json
import subprocess
from typing import List, Dict, Any, Optional, TypedDict

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

def get_github_token(environ=os.environ, subprocess_module=subprocess) -> str:
    """
    Retrieve a GitHub token using multiple methods.
    
    1. Check GITHUB_TOKEN environment variable
    2. Try to get token using GitHub CLI
    3. Raise AuthenticationError if all methods fail
    
    Args:
        environ: Environment dictionary to use (default: os.environ)
        subprocess_module: Subprocess module to use (default: subprocess)
    
    Returns:
        GitHub token as string
        
    Raises:
        AuthenticationError: If unable to retrieve a valid GitHub token
    """
    # Method 1: Environment variable
    token = environ.get('GITHUB_TOKEN')
    if token:
        return token
    
    # Method 2: GitHub CLI
    try:
        result = subprocess_module.run(
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
    error_msg = (
        "Unable to retrieve GitHub token.\n"
        "Please either:\n"
        "  1. Set the GITHUB_TOKEN environment variable, or\n"
        "  2. Install and authenticate with GitHub CLI (gh)"
    )
    raise AuthenticationError(error_msg)

def list_versions(namespace: str, package_name: str, token_func=get_github_token) -> List[PackageVersion]:
    """
    List all versions of a package in the GitHub Container Registry.
    
    Args:
        namespace: The namespace in form of 'user/<username>' or 'org/<orgname>'
        package_name: The name of the package
        token_func: Function that returns a GitHub token (default: get_github_token)
    
    Returns:
        List of package versions with metadata
        
    Raises:
        AuthenticationError: If authentication fails
        APIError: If API request fails
    """
    # Get authentication token
    token = token_func()
    
    # Construct API URL
    api_url = f"https://api.github.com/{namespace}/packages/container/{package_name}/versions?per_page=100"
    
    # Set up headers with authentication
    headers = {
        "Accept": "application/vnd.github+json",
        "Authorization": f"Bearer {token}",
        "X-GitHub-Api-Version": "2022-11-28"
    }
    
    response = requests.get(api_url, headers=headers)
    response.raise_for_status()
    return response.json()

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
    except Exception as e:
        print(f"Command failed: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
