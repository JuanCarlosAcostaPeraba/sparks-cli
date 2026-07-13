package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/model"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/presentation"
)

type Renderer struct {
	w       io.Writer
	palette presentation.Palette
}

func NewRenderer(w io.Writer, color bool) Renderer {
	return Renderer{w: w, palette: presentation.Palette{Enabled: color}}
}

func rendererFor(w io.Writer) Renderer {
	return Renderer{w: w, palette: presentation.ForWriter(w)}
}

func Sparks(w io.Writer, sparks []model.Spark, asJSON bool) error {
	return rendererFor(w).Sparks(sparks, asJSON)
}

func (r Renderer) Sparks(sparks []model.Spark, asJSON bool) error {
	if asJSON {
		enc := json.NewEncoder(r.w)
		enc.SetIndent("", "  ")
		return enc.Encode(sparks)
	}
	return r.table(sparks)
}

func (r Renderer) table(sparks []model.Spark) error {
	var plain bytes.Buffer
	tw := tabwriter.NewWriter(&plain, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "STATUS\tID\tTITLE"); err != nil {
		return err
	}
	for _, spark := range sparks {
		if _, err := fmt.Fprintf(tw, "%s\t%d\t%s\n", symbol(spark), spark.ID, spark.Title); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	if !r.palette.Enabled {
		_, err := io.Copy(r.w, &plain)
		return err
	}

	lines := strings.Split(strings.TrimSuffix(plain.String(), "\n"), "\n")
	if len(lines) > 0 {
		lines[0] = r.palette.Paint(presentation.Muted, lines[0])
	}
	for index, spark := range sparks {
		lines[index+1] = r.colorTableLine(lines[index+1], spark)
	}
	_, err := fmt.Fprintln(r.w, strings.Join(lines, "\n"))
	return err
}

func (r Renderer) colorTableLine(line string, spark model.Spark) string {
	status := symbol(spark)
	id := strconv.FormatInt(spark.ID, 10)
	idStart := strings.Index(line[len(status):], id)
	if idStart < 0 {
		return line
	}
	idStart += len(status)
	titleStart := strings.LastIndex(line, spark.Title)
	if titleStart < idStart+len(id) {
		return line
	}

	role := presentation.Muted
	if spark.Important {
		role = presentation.Important
	} else if spark.Done {
		role = presentation.Completed
	}

	var result strings.Builder
	result.WriteString(r.palette.Paint(role, status))
	result.WriteString(line[len(status):idStart])
	result.WriteString(r.palette.Paint(presentation.ID, id))
	result.WriteString(line[idStart+len(id) : titleStart])
	if spark.Important || spark.Done {
		result.WriteString(r.palette.Paint(role, spark.Title))
	} else {
		result.WriteString(spark.Title)
	}
	return result.String()
}

func Tree(w io.Writer, sparks []model.Spark, asJSON bool) error {
	return rendererFor(w).Tree(sparks, asJSON)
}

func (r Renderer) Tree(sparks []model.Spark, asJSON bool) error {
	if asJSON {
		return r.Sparks(sparks, true)
	}

	byParent := map[int64][]model.Spark{}
	var roots []model.Spark
	for _, spark := range sparks {
		if spark.ParentID == nil {
			roots = append(roots, spark)
			continue
		}
		byParent[*spark.ParentID] = append(byParent[*spark.ParentID], spark)
	}

	var walk func(items []model.Spark, prefix string, parentNumber string)
	walk = func(items []model.Spark, prefix string, parentNumber string) {
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].ID < items[j].ID
		})
		for i, spark := range items {
			number := treeNumber(parentNumber, i+1)
			connector := "├─"
			nextPrefix := prefix + "│  "
			if i == len(items)-1 {
				connector = "└─"
				nextPrefix = prefix + "   "
			}
			role := presentation.Muted
			if spark.Important {
				role = presentation.Important
			} else if spark.Done {
				role = presentation.Completed
			}
			title := spark.Title
			if spark.Important || spark.Done {
				title = r.palette.Paint(role, title)
			}
			fmt.Fprintf(r.w, "%s%s %s %s) %s\n",
				r.palette.Paint(presentation.Muted, prefix),
				r.palette.Paint(presentation.Muted, connector),
				r.palette.Paint(role, symbol(spark)),
				r.palette.Paint(presentation.ID, number),
				title,
			)
			walk(byParent[spark.ID], nextPrefix, number)
		}
	}

	if len(roots) == 0 {
		return nil
	}
	walk(roots, "", "")
	return nil
}

func treeNumber(parent string, position int) string {
	current := strconv.Itoa(position)
	if parent == "" {
		return current
	}
	return parent + "." + current
}

func Message(w io.Writer, format string, args ...any) {
	rendererFor(w).Message(format, args...)
}

func (r Renderer) Message(format string, args ...any) {
	message := fmt.Sprintf(strings.TrimRight(format, "\n"), args...)
	if r.palette.Enabled {
		fmt.Fprintf(r.w, "%s %s\n", r.palette.Paint(presentation.Success, "✓"), message)
		return
	}
	fmt.Fprintln(r.w, message)
}

func ID(w io.Writer, value any) string {
	return rendererFor(w).ID(value)
}

func (r Renderer) ID(value any) string {
	return r.palette.Paint(presentation.ID, fmt.Sprint(value))
}

func symbol(spark model.Spark) string {
	switch {
	case spark.Done:
		return "☑"
	case spark.Important:
		return "❗"
	default:
		return "□"
	}
}
