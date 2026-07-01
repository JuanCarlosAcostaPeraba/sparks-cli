package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/model"
)

func Sparks(w io.Writer, sparks []model.Spark, asJSON bool) error {
	if asJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(sparks)
	}

	for _, spark := range sparks {
		fmt.Fprintf(w, "%s %d) %s\n", symbol(spark), spark.ID, spark.Title)
	}
	return nil
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

	var walk func(items []model.Spark, prefix string)
	walk = func(items []model.Spark, prefix string) {
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].ID < items[j].ID
		})
		for i, spark := range items {
			connector := "├─"
			nextPrefix := prefix + "│  "
			if i == len(items)-1 {
				connector = "└─"
				nextPrefix = prefix + "   "
			}
			fmt.Fprintf(w, "%s%s %s %d) %s\n", prefix, connector, symbol(spark), spark.ID, spark.Title)
			walk(byParent[spark.ID], nextPrefix)
		}
	}

	if len(roots) == 0 {
		return nil
	}
	walk(roots, "")
	return nil
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
