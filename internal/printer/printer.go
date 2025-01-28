package printer

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"gitlab.com/slon/shad-go/gitfame/internal/flag"
	"gitlab.com/slon/shad-go/gitfame/internal/statistic"
)

func Print(stats *statistic.Statistic, flag flag.Flag) {
	switch flag.Format {
	case "tabular":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintf(w, "Name\tLines\tCommits\tFiles\n")
		for _, author := range stats.GetSorted(flag) {
			fmt.Fprintf(w, "%s\t%d\t%d\t%d\n", author.Name, author.Lines, author.Commits(), author.Files)
		}
		w.Flush()
	case "csv":
		w := csv.NewWriter(os.Stdout)
		err := w.Write([]string{"Name", "Lines", "Commits", "Files"})
		if err != nil {
			panic(err)
		}
		for _, author := range stats.GetSorted(flag) {
			err := w.Write([]string{author.Name,
				strconv.Itoa(author.Lines),
				strconv.Itoa(author.Commits()),
				strconv.Itoa(author.Files)})
			if err != nil {
				panic(err)
			}
		}
		w.Flush()
	case "json":
		jsonData, err := json.Marshal(stats.GetSorted(flag))
		if err != nil {
			panic(err)
		}

		fmt.Println(string(jsonData))
	case "json-lines":
		for _, author := range stats.GetSorted(flag) {
			jsonData, err := author.MarshalJSON()
			if err != nil {
				panic(err)
			}
			fmt.Println(string(jsonData))
		}
	}
}
