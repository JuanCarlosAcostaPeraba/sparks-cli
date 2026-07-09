package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/model"
)

func Sparks(w io.Writer, sparks []model.Spark, asJSON bool) error {
	if asJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(sparks)
	}

	return table(w, sparks)
}

func table(w io.Writer, sparks []model.Spark) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "STATUS\tID\tTITLE"); err != nil {
		return err
	}
	for _, spark := range sparks {
		if _, err := fmt.Fprintf(tw, "%s\t%d\t%s\n", symbol(spark), spark.ID, spark.Title); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func Tree(w io.Writer, sparks []model.Spark, asJSON bool) error {
	if asJSON {
		return Sparks(w, sparks, true)
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
			fmt.Fprintf(w, "%s%s %s %s) %s\n", prefix, connector, symbol(spark), number, spark.Title)
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
	fmt.Fprintf(w, strings.TrimRight(format, "\n")+"\n", args...)
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
