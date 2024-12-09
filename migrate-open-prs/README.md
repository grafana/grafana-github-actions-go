# migrate-open-prs

This GitHub Action automatically migrates open pull requests from a superseded release branch to the current release branch and notifies the PR authors of the changes. It's particularly useful during version transitions when you need to ensure that ongoing work targets the latest release branch.

The action performs several key tasks:

1. Identifies all open PRs targeting the specified previous release branch
2. Updates each PR's base branch to target the new release branch
3. Notifies PR authors about the migration, including whether it was successful or if manual intervention is needed

## Inputs

| Input        | Description                                                           | Required |
| ------------ | --------------------------------------------------------------------- | -------- |
| `prevBranch` | The superseded release branch to check for open PRs (e.g., `v10.0.x`) | Yes      |
| `nextBranch` | The current release branch to migrate open PRs to (e.g., `v10.1.x`)   | Yes      |

## Environment Variables

| Variable       | Description                         | Required |
| -------------- | ----------------------------------- | -------- |
| `GITHUB_TOKEN` | GitHub token for API authentication | Yes      |
