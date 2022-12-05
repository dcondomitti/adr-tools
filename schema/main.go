package schema

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Decision struct {
	GithubURL  string
	Path       string
	RawContent string
	Title      string
	Number     int
	Status     State
	Date       time.Time
}

type State int

const (
	proposed State = iota
	accepted
	rejected
	superseded
)

var titleMatcher = regexp.MustCompile(`\A# \d+. (.*)`)
var statusMatcher = regexp.MustCompile(`## Status\s*(.*)`)

func (d Decision) TitleFromDocument() string {
	titles := titleMatcher.FindStringSubmatch(d.RawContent)
	if len(titles) <= 1 {
		return "Untitled document"
	}

	return titles[1]
}

func (d Decision) Filename() string {
	return filepath.Base(d.Path)
}

func (d Decision) ShortID() string {
	parts := strings.Split(d.Filename(), "-")
	if len(parts) < 2 {
		return "0000"
	}
	return parts[0]
}

func (d Decision) ID() string {
	parts := strings.Split(d.Filename(), "-")
	if len(parts) < 2 {
		return "ADR-0000"
	}
	return fmt.Sprintf("ADR-%s", parts[0])
}

func (d Decision) StatusFromDocument() string {
	statuses := statusMatcher.FindStringSubmatch(d.RawContent)
	if len(statuses) <= 1 {
		return "Unknown"
	}

	return statuses[1]
}

func (d Decision) IsActive() bool {
	if strings.Contains(d.StatusFromDocument(), "Superseded") {
		return false
	}

	return true
}
