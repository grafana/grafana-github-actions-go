# Create Security Branch Action

This GitHub Action creates a new security branch in a private repository for handling security fixes. It's designed to work with Grafana's security release process.

## Usage

```yaml
- name: Create security branch
  uses: grafana/grafana-github-actions-go/create-security-branch@main
  with:
    version: '12.0.1'  # The version to create a security branch for
    security_branch_number: '01'  # The security branch number (two digits)
    repository: 'grafana/grafana-security-mirror'  # The target repository
    token: ${{ secrets.GITHUB_TOKEN }}  # GitHub token with repository access
```

## Inputs

| Name | Description | Required | Default |
|------|-------------|----------|---------|
| `version` | The version to create a security branch for (e.g., 12.0.1) | Yes | - |
| `security_branch_number` | The security branch number (e.g., 01) | Yes | - |
| `repository` | The repository to create the security branch in (e.g., grafana/grafana-security-mirror) | Yes | - |
| `token` | GitHub token with access to the target repository | Yes | - |

## Outputs

| Name | Description |
|------|-------------|
| `branch` | The name of the created security branch |

## Example

When creating a security branch for version 12.0.1 with security branch number 01, the action will:
1. Create a branch named `12.0.1+security-01`
2. Base it on the `release-12.0.1` branch
3. Push it to the specified repository
