import sys
import os
import unittest
from unittest.mock import MagicMock

# Add the parent directory to sys.path to import the ghcr module
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))
from ghcr import get_github_token, list_versions, AuthenticationError


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
        
        # Sample API response
        mock_response = MagicMock()
        mock_response.json.return_value = [
            {
                "id": 1234,
                "name": "1.0.0",
                "url": "https://api.github.com/user/test/packages/container/test-package/versions/1234",
                "package_html_url": "https://github.com/user/test/packages/container/package/test-package",
                "created_at": "2023-01-01T00:00:00Z",
                "updated_at": "2023-01-01T00:00:00Z",
                "html_url": "https://github.com/user/test/packages/container/test-package/1234",
                "metadata": {
                    "container": {
                        "tags": ["latest", "1.0.0"]
                    }
                }
            }
        ]
        
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
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0]["id"], 1234)
        self.assertEqual(result[0]["name"], "1.0.0")
        self.assertEqual(result[0]["metadata"]["container"]["tags"], ["latest", "1.0.0"])

if __name__ == "__main__":
    unittest.main()
