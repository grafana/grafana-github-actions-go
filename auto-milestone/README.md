# auto-milestone

This action is usually triggered right after merging a pull request in order to set the correct milestone for it.
The milestone is determined by the Grafana version of the branch the PR is merged into.

You can also dry-run it using the following command:

```
$ go run ./auto-milestone --repo grafana/grafana $PR_NUMBER
```

## Example workflow:

```yaml
name: Auto-milestone
on:
  pull_request:
    types:
      - closed
jobs:
  main:
    runs-on: ubuntu-latest
    steps:
      - name: Run auto-milestone
        uses: grafana/grafana-github-actions-go/auto-milestone@main
        with:
          pr: ${{ github.event.pull_request.number }}
          token: ${{ secrets.GITHUB_TOKEN }}

```
