package formatters

import (
	"encoding/json"
	"fmt"
	"io"
	"jnsgruk/ghstat/internal/greenhouse"
	"log/slog"
	"strconv"

	"github.com/fatih/color"
	"github.com/fbiville/markdown-table-formatter/pkg/markdown"
	"github.com/rodaine/table"
)

// Formatter interface is a generic interface for an ghstat output format
type Formatter interface {
	Output(roles []*greenhouse.Role)
}

// NewFormatter constructs a formatter of the requested type
func NewFormatter(input string, writer io.Writer) Formatter {
	switch input {
	case "pretty":
		return &PrettyTableFormatter{writer: writer}
	case "markdown":
		return &MarkdownTableFormatter{writer: writer}
	case "json":
		return &JsonFormatter{writer: writer}
	default:
		return nil
	}
}

// JsonFormatter is a simple formatter that marshals the gathered information
// about a set of roles to a simple json format
type JsonFormatter struct {
	writer io.Writer
}

// Output dumps the role information to stdout as JSON
func (o *JsonFormatter) Output(roles []*greenhouse.Role) {
	b, err := json.MarshalIndent(roles, "", "  ")
	if err != nil {
		slog.Error("could not marshal output data", "error", err.Error())
	}
	fmt.Fprint(o.writer, string(b))
}

// MarkdownTableFormatter is used for rendering stats as a Markdown table
type MarkdownTableFormatter struct {
	writer io.Writer
}

// Output dumps the role information as a Markdown table to stdout
func (o *MarkdownTableFormatter) Output(roles []*greenhouse.Role) {
	rows := [][]string{}
	for _, r := range roles {
		rows = append(rows, []string{
			r.Lead,
			r.Title,
			strconv.Itoa(r.AppReviews()),
			strconv.Itoa(r.NeedsDecision()),
			strconv.Itoa(r.NeedsScheduling()),
			strconv.Itoa(r.WIScreening()),
			strconv.Itoa(r.WIGrading()),
			strconv.Itoa(r.Stale()),
		})
	}

	tbl, _ := markdown.NewTableFormatterBuilder().
		WithPrettyPrint().
		Build("Lead", "Role", "CVs", "Decisions", "Scheduling", "WI (Screen)", "WI (Grade)", "Stale").
		Format(rows)

	fmt.Fprint(o.writer, tbl)
}

// PrettyTableFormatter dumps the role information to a pretty printed terminal
type PrettyTableFormatter struct {
	writer io.Writer
}

// Output dumps the pretty table to stdout
func (o *PrettyTableFormatter) Output(roles []*greenhouse.Role) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("Lead", "Role", "CVs", "Decisions", "Scheduling", "WI (Screen)", "WI (Grade)", "Stale").WithWriter(o.writer)
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, r := range roles {
		tbl.AddRow(
			r.Lead,
			r.Title,
			r.AppReviews(),
			r.NeedsDecision(),
			r.NeedsScheduling(),
			r.WIScreening(),
			r.WIGrading(),
			r.Stale(),
		)
	}
	tbl.Print()
}
