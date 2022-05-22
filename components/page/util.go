package page

import (
	"strings"
	"wander/formatter"
)

type pageData struct {
	allData, filteredData []string
}

func dataAsTable(data []string, columnTitles []string) formatter.Table {
	var splitData [][]string
	for _, row := range data {
		splitData = append(splitData, strings.Fields(row))
	}
	return formatter.GetRenderedTableAsString(columnTitles, splitData)
}
