# community-release

This action allows you to quickly generate a new release post on https://community.grafana.com/ based on the changelog for the provided version.

You can also dry-run it using the following command:

```
$ INPUT_VERSION=... go run ./community-release --preview --repo grafana/grafana
# e.g. INPUT_VERSION=9.4.12 go run ./community-release --preview --repo grafana/grafana
```

## Example workflow:

```
name: Create or update GitHub release
on:
  workflow_dispatch:
    inputs:
      version:
        required: true
        description: Needs to match, exactly, the name of a milestone (NO v prefix)
jobs:
  main:
    runs-on: ubuntu-latest
    steps:
      - name: Run github release action
        uses: grafana/grafana-github-actions-go/community-release@main
        with:
          version: ${{ inputs.version }}
          token: ${{ secrets.GITHUB_TOKEN }}
          metricsWriteAPIKey: ${{secrets.GRAFANA_MISC_STATS_API_KEY}}
          community_api_key: ${{ secrets.COMMUNITY_API_KEY }}
          community_api_username:${{ secrets.COMMUNITY_USERNAME }}
```
