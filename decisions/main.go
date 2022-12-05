package decisions

import (
	"bytes"
	_ "embed"
	"text/template"
	"time"

	"github.com/bmorton/adr-tools/schema"
	"github.com/k0kubun/pp"
)

//go:embed templates/decision.md.tmpl
var decisionTemplate string

type Builder struct {
	schema.Decision
}

func NewBuilder(title string) *Builder {
	return &Builder{
		schema.Decision{
			Title:  title,
			Date:   time.Now(),
			Status: schema.State,
		},
	}
}

func (d Builder) GenerateDecision() (string, error) {
	t := template.New("decision.md")
	t = template.Must(t.Parse(decisionTemplate))

	buf := bytes.NewBufferString("")
	err := t.Execute(buf, schema.Decision{
		Title: d.Decision.Title,
		Date:  d.Decision.Date,
	})

	pp.Println(buf.String())
	buf.Reset()
	return buf.String(), err
}
