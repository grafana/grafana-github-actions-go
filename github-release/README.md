# github-release

This action allows you to quickly generate a new GitHub released based on the provided version.
The underlying tag with the same name as the version needs to exist.
The content of the release will be generated based on the *existing* changelog.

You can also dry-run it using the following command:

```
$ go run ./github-release --preview --repo grafana/grafana $VERSION
# e.g. go run ./github-release --preview --repo grafana/grafana 9.4.12
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
      latest:
        required: false
        description: Mark the new release as latest (`1` or `0`)
jobs:
  main:
    runs-on: ubuntu-latest
    steps:
      - name: Run github release action
        uses: grafana/grafana-github-actions-go/auto-milestone@main
        with:
          token: ${{ secrets.GITHUB_TOKEN }}
          metricsWriteAPIKey: ${{secrets.GRAFANA_MISC_STATS_API_KEY}}

```
