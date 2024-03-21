set -euo pipefail

# check if there is a diff in the .changelog directory
# for PRs against the main branch, the changelog file name should match the PR number
if [ "$GITHUB_BASE_REF" = "$GITHUB_DEFAULT_BRANCH" ]; then
    enforce_matching_pull_request_number="matching this PR number "
    changelog_file_path=".changelog/(_)?$PR_NUMBER.txt"
else
    changelog_file_path=".changelog/[_0-9]*.txt"
fi

changelog_files=$(git --no-pager diff --name-only HEAD "$(git merge-base HEAD "origin/main")" | egrep "${changelog_file_path}")

# If we do not find a file in .changelog/, we fail the check
if [ -z "$changelog_files" ]; then
    # Fail status check when no .changelog entry was found on the PR
    echo "Did not find a .changelog entry ${enforce_matching_pull_request_number}and the 'pr/no-changelog' label was not applied."
    exit 1
fi

# Install the changelog-check command
go install github.com/hashicorp/go-changelog/cmd/changelog-check@latest

# Validate format with make changelog-check, exit with error if any note has an
# invalid format
for file in $changelog_files; do
  if ! cat $file | make changelog/check; then
    echo "Found a changelog entry ${enforce_matching_pull_request_number}but the note format in ${file} was invalid."
    exit 1
  fi
done

echo "Found valid changelog entry!"
