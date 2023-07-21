# github-release

This action allows you to quickly generate a new GitHub released based on the provided version.
The underlying tag with the same name as the version needs to exist.
The content of the release will be generated based on the *existing* changelog.

You can also dry-run it using the following command:

```
$ go run ./github-release --preview --repo grafana/grafana $VERSION
# e.g. go run ./github-release --preview --repo grafana/grafana 9.4.12
```

