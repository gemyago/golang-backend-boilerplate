import sys
import os
from typing import List
import unittest
import random
import re
from unittest.mock import MagicMock, patch
from faker import Faker
from datetime import datetime, timedelta

import requests

fake = Faker()

# Add the parent directory to sys.path to import the ghcr module
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))
from ghcr import (
    get_github_token, 
    list_versions, 
    find_versions_to_clean,
    remove_version,
    AuthenticationError, 
    PackageVersion,
    CleanupAction
)


def create_random_package_version(**overrides) -> PackageVersion:
    """
    Create a random PackageVersion object for testing purposes.
    
    Args:
        overrides: Optional keyword arguments to override default values

    Returns:
        A randomly generated PackageVersion object
    """
    
    version = {
        "id": fake.random_int(min=1000, max=9999),
        "name": fake.name(),
        "url": fake.uri(),
        "package_html_url": fake.uri(),
        "created_at": fake.past_datetime().isoformat().replace('+00:00', 'Z'),
        "updated_at": fake.past_datetime().isoformat().replace('+00:00', 'Z'),
        "html_url": fake.uri(),
        "metadata": {
            "container": {
                "tags": fake.words()
            }
        }
    }
    version.update(overrides)
    return PackageVersion(**version)

def create_dated_package_versions(num_versions, date_pattern='recent') -> List[PackageVersion]:
    """
    Create a list of package versions with specific dating patterns for testing.
    
    Args:
        num_versions: Number of versions to create
        date_pattern: 'recent' for versions created recently, 'sequential' for 
                     versions with sequential dates
    
    Returns:
        List of versioned packages with controlled dates
    """
    versions = []
    base_date = datetime.now()
    
    for i in range(num_versions):
        if date_pattern == 'recent':
            # Create some versions from today, some from last week, some from last month
            if i < num_versions // 3:
                days_ago = random.randint(0, 2)  # Last couple days
            elif i < 2 * (num_versions // 3):
                days_ago = random.randint(3, 10)  # Last week or so
            else:
                days_ago = random.randint(20, 60)  # Last month or two
        else:  # sequential
            # Each version is created 1 day before the previous one
            days_ago = i
            
        created_date = (base_date - timedelta(days=days_ago)).isoformat().replace('+00:00', 'Z')
        
        version = create_random_package_version(
            id=1000 + i,
            created_at=created_date,
            updated_at=created_date
        )

        versions.append(version)
    
    return versions


class TestGetGitHubToken(unittest.TestCase):
    def test_get_token_from_env(self):
        """Test getting token from environment variable"""
        mock_environ = {'GITHUB_TOKEN': 'test-token-from-env'}
        
        # Mock subprocess to ensure it's not used when env var is available
        mock_subprocess = MagicMock()
        
        token = get_github_token(environ=mock_environ, subprocess_module=mock_subprocess)
        
        self.assertEqual(token, 'test-token-from-env')
        # Verify subprocess was not called
        mock_subprocess.run.assert_not_called()
    
    def test_get_token_from_gh_cli(self):
        """Test getting token from GitHub CLI when env var is not available"""
        # Empty environment - no GITHUB_TOKEN
        mock_environ = {}
        
        # Mock successful subprocess response
        mock_process = MagicMock()
        mock_process.returncode = 0
        mock_process.stdout = "test-token-from-gh-cli\n"
        
        mock_subprocess = MagicMock()
        mock_subprocess.run.return_value = mock_process
        
        token = get_github_token(environ=mock_environ, subprocess_module=mock_subprocess)
        
        self.assertEqual(token, 'test-token-from-gh-cli')
        # Verify subprocess was called with correct arguments
        mock_subprocess.run.assert_called_once_with(
            ["gh", "auth", "token"],
            capture_output=True,
            text=True,
            check=False
        )
    
    def test_gh_cli_not_found(self):
        """Test handling when GitHub CLI is not installed"""
        # Empty environment - no GITHUB_TOKEN
        mock_environ = {}
        
        # Mock subprocess that raises FileNotFoundError (gh not installed)
        mock_subprocess = MagicMock()
        mock_subprocess.run.side_effect = FileNotFoundError("No such file or directory: 'gh'")
        
        # Should raise AuthenticationError
        with self.assertRaises(AuthenticationError) as context:
            get_github_token(environ=mock_environ, subprocess_module=mock_subprocess)
        
        # Verify error message contains helpful instructions
        error_message = str(context.exception)
        self.assertIn("Unable to retrieve GitHub token", error_message)
        self.assertIn("Set the GITHUB_TOKEN environment variable", error_message)
        self.assertIn("Install and authenticate with GitHub CLI", error_message)
    
    def test_gh_cli_error(self):
        """Test handling when GitHub CLI returns an error"""
        # Empty environment - no GITHUB_TOKEN
        mock_environ = {}
        
        # Mock failed subprocess response
        mock_process = MagicMock()
        mock_process.returncode = 1
        mock_process.stdout = ""
        
        mock_subprocess = MagicMock()
        mock_subprocess.run.return_value = mock_process
        
        # Should raise AuthenticationError
        with self.assertRaises(AuthenticationError) as context:
            get_github_token(environ=mock_environ, subprocess_module=mock_subprocess)
        
        # Verify error message contains helpful instructions
        error_message = str(context.exception)
        self.assertIn("Unable to retrieve GitHub token", error_message)


class TestListVersions(unittest.TestCase):
    def test_list_versions_success(self):
        """Test listing versions with successful API response"""
        # Mock token function
        mock_token_func = MagicMock(return_value="test-token")
        
        # Generate random package versions
        num_versions = random.randint(3, 5)
        mock_versions = [create_random_package_version() for _ in range(num_versions)]
        
        # Sample API response
        mock_response = MagicMock()
        mock_response.json.return_value = mock_versions
        
        # Mock requests module
        mock_requests = MagicMock()
        mock_requests.get.return_value = mock_response
        
        # Call the function
        result = list_versions(
            namespace="user/test",
            package_name="test-package",
            token_func=mock_token_func,
            requests_module=mock_requests
        )
        
        # Verify API was called correctly
        mock_requests.get.assert_called_once_with(
            "https://api.github.com/user/test/packages/container/test-package/versions?per_page=100",
            headers={
                "Accept": "application/vnd.github+json",
                "Authorization": "Bearer test-token",
                "X-GitHub-Api-Version": "2022-11-28"
            }
        )
        
        # Verify the response was parsed correctly
        self.assertEqual(len(result), num_versions)
        # Verify each result matches the original mock data
        for i, version in enumerate(result):
            self.assertEqual(version["id"], mock_versions[i]["id"])
            self.assertEqual(version["name"], mock_versions[i]["name"])
            self.assertEqual(version["metadata"]["container"]["tags"], 
                             mock_versions[i]["metadata"]["container"]["tags"])


class TestFindVersionsToClean(unittest.TestCase):
    def test_keep_latest_versions(self):
        """Test keeping the latest N versions by date"""
        versions = create_dated_package_versions(10, date_pattern='sequential')
        
        cleanup_actions = find_versions_to_clean(versions, keep_latest=3)
        
        # Extract versions marked for deletion
        to_remove = [action for action in cleanup_actions if action["action"] == "delete"]
        # Extract versions marked to keep
        to_keep = [action for action in cleanup_actions if action["action"] == "keep"]
        
        # Verify counts
        self.assertEqual(len(to_remove), 7)
        self.assertEqual(len(to_keep), 3)
        
        # Verify the correct versions are kept (most recent ones)
        kept_ids = {v["version"]['id'] for v in to_keep}
        self.assertEqual(kept_ids, {versions[0]['id'], versions[1]['id'], versions[2]['id'], })
        
        # Verify all actions have a valid reason
        for action in to_keep:
            self.assertRegex(action["reason"], r"most recent")

        for action in to_remove:
            self.assertRegex(action["reason"], r"Older tagged")

    def test_empty_versions_list(self):
        """Test with an empty list of versions"""
        cleanup_actions = find_versions_to_clean([])
        self.assertEqual(len(cleanup_actions), 0)
    
    def test_find_untagged_versions(self):
        """Test finding versions with no tags"""
        with_tags = [
            create_random_package_version(metadata={"container": {"tags": ["tag1", "latest"]}})
            for _ in range(6)
        ]

        without_tags = [
            create_random_package_version(metadata={"container": {"tags": []}}) 
            for _ in range(6)
        ]
        
        cleanup_actions = find_versions_to_clean(with_tags + without_tags, keep_latest=len(with_tags))
        
        # All untagged versions should be marked for deletion
        self.assertEqual(len(cleanup_actions), len(with_tags) + len(without_tags))

        to_remove = [action for action in cleanup_actions if action["action"] == "delete"]
        self.assertEqual({a["version"]["id"] for a in to_remove}, {a["id"] for a in without_tags})
        
        # Verify the reason for deletion
        for action in cleanup_actions:
            if action["action"] == "delete":
                self.assertEqual(action["reason"], "Untagged version")
    
    def test_keep_tagged_versions(self):
        """Test keeping versions with tags"""
        versions = []
        
        for i in range(5):
            version = create_random_package_version(
                id=1000 + i,
                metadata={"container": {"tags": [f"tag-{i}", "latest"]}}
            )
            versions.append(version)
        
        cleanup_actions = find_versions_to_clean(versions)
        
        # Extract versions marked for deletion
        to_remove = [action["version"] for action in cleanup_actions if action["action"] == "delete"]
        
        # No versions should be marked for deletion (all 5 are tagged and within keep_latest=5)
        self.assertEqual(len(to_remove), 0)

class TestRemoveVersion(unittest.TestCase):
    def test_remove_version_dry_run(self):
        """Test removing a version in dry run mode"""
        # Mock token function
        mock_token_func = MagicMock(return_value="test-token")
        
        # Mock requests module
        mock_requests = MagicMock()
        
        # Call remove_version with dry_run=True
        result = remove_version(
            namespace="user/test",
            package_name="test-package",
            version_id=1234,
            dry_run=True,
            token_func=mock_token_func,
            requests_module=mock_requests
        )
        
        # Verify result is True
        self.assertTrue(result)
        
        # Verify no API calls were made
        mock_requests.delete.assert_not_called()
    
    def test_remove_version_actual(self):
        """Test removing a version for real"""
        # Mock token function
        mock_token_func = MagicMock(return_value="test-token")
        
        # Mock requests module
        mock_response = MagicMock()
        mock_requests = MagicMock()
        mock_requests.delete.return_value = mock_response
        
        # Call remove_version with dry_run=False
        result = remove_version(
            namespace="user/test",
            package_name="test-package",
            version_id=1234,
            dry_run=False,
            token_func=mock_token_func,
            requests_module=mock_requests
        )
        
        # Verify result is True
        self.assertTrue(result)
        
        # Verify API was called correctly
        mock_requests.delete.assert_called_once_with(
            "https://api.github.com/user/test/packages/container/test-package/versions/1234",
            headers={
                "Accept": "application/vnd.github+json",
                "Authorization": "Bearer test-token",
                "X-GitHub-Api-Version": "2022-11-28"
            }
        )
        
    def test_remove_version_error(self):
        """Test error handling when removing a version"""
        # Mock token function
        mock_token_func = MagicMock(return_value="test-token")
        
        # Mock requests module with an error response
        mock_response = MagicMock()
        mock_response.raise_for_status.side_effect = requests.exceptions.HTTPError("API Error")
        mock_requests = MagicMock()
        mock_requests.delete.return_value = mock_response
        
        # Call remove_version with dry_run=False
        with self.assertRaises(requests.exceptions.HTTPError):
            remove_version(
                namespace="user/test",
                package_name="test-package",
                version_id=1234,
                dry_run=False,
                token_func=mock_token_func,
                requests_module=mock_requests
            )

if __name__ == "__main__":
    unittest.main()
