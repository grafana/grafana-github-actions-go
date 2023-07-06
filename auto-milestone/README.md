# auto-milestone

This action is usually triggered right after merging a pull request in order to set the correct milestone for it.
The milestone is determined by the Grafana version of the branch the PR is merged into.

You can also dry-run it using the following command:

```
$ go run ./auto-milestone --repository grafana/grafana $PR_NUMBER
```
