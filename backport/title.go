package main

import "strings"

const defaultTitleTemplate = "[{{branch}}] {{title}}"

func FormatBackportTitle(template, branch, title string) string {
	if template == "" {
		template = defaultTitleTemplate
	}
	return strings.NewReplacer(
		"{{branch}}", branch,
		"{{title}}", title,
	).Replace(template)
}
