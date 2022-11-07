package schema

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type Decision struct {
	Path       string
	RawContent string
	GithubURL  string
}

var titleMatcher = regexp.MustCompile(`\A# \d+. (.*)`)
var statusMatcher = regexp.MustCompile(`## Status\s*(.*)`)

func (d Decision) Title() string {
	titles := titleMatcher.FindStringSubmatch(d.RawContent)
	if len(titles) <= 1 {
		return "Untitled document"
	}

	return titles[1]
}

func (d Decision) Filename() string {
	return filepath.Base(d.Path)
}

func (d Decision) ID() string {
	parts := strings.Split(d.Filename(), "-")
	if len(parts) == 0 {
		return "ADR-0000"
	}
	return fmt.Sprintf("ADR-%s", parts[0])
}

func (d Decision) Status() string {
	statuses := statusMatcher.FindStringSubmatch(d.RawContent)
	if len(statuses) <= 1 {
		return "Unknown"
	}

	return statuses[1]
}
