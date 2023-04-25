# grafana-github-actions-go

This repository contains a handful of GitHub Actions that are mostly used in the context of releasing Grafana packages.

## update-changelog

This action generates an update to the `CHANGELOG.md` file based on pull-requests included in a release-milestone.
The following inputs need to be provided:

- `version` (e.g. `9.4.0`) which matches the name of the relevant milestone inside the grafana/grafana project.
- `token` which represents a GitHub token which has read access to the relevant projects *and* can push to the target project for creating a new PR with the updated changelog.
- `metrics_api_endpoint` (default: `https://graphite-us-central1.grafana.net/metrics`): Graphite HTTP endpoint to submit usage metrics to.
- `metrics_api_key`: API key for that Graphite endpoint (will be used as HTTP Basic Auth password).
- `metrics_api_username`: Username for that Graphite endpoint.
- `community_api_key`: API key to be used for API calls to the community.
- `community_api_username`: Username to be used for API calls to the community.
- `community_base_url`: Base URL of the community forums.
- `community_category_id`: Category where to post the changelog to in the community.

Example workflow:

```yaml
name: Update changelog
on:
  workflow_dispatch:
    inputs:
      version:
        description: Needs to match, exactly, the name of a version
        required: true
        type: string
jobs:
  main:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: grafana/grafana-github-actions-go/update-changelog@main
        with:
          version: ${{ inputs.version }}
          token: "${{secrets.GH_TOKEN}}"
```
