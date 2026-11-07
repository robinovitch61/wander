package page

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// SortValue is what a table cell sorts by: its meaning rather than its rendered text,
// e.g. a time as nanoseconds, or a count as its running then wanted numbers
type SortValue struct {
	Text    string
	Nums    []float64
	Missing bool
}

// TextSortValue sorts case-insensitively, with runs of digits compared as numbers, so "2" comes
// before "10" and "1.9" before "1.10". A column whose text is all plain decimal numbers sorts by
// their values instead (see sortRows). Empty and "-" are missing.
func TextSortValue(s string) SortValue {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" || trimmed == "-" {
		return MissingSortValue()
	}
	return SortValue{Text: strings.ToLower(trimmed)}
}

// NumSortValue sorts by its numbers in turn, the later ones breaking ties of the earlier
func NumSortValue(first float64, rest ...float64) SortValue {
	return SortValue{Nums: append([]float64{first}, rest...)}
}

func MissingSortValue() SortValue {
	return SortValue{Missing: true}
}

// compareSortValues returns -1, 0 or 1. Numbers come before text. Missing values compare equal here;
// sortRows keeps them last in either direction.
func compareSortValues(a, b SortValue) int {
	switch {
	case a.Nums != nil && b.Nums != nil:
		return slices.Compare(a.Nums, b.Nums)
	case a.Nums != nil:
		return -1
	case b.Nums != nil:
		return 1
	}
	return compareNaturally(a.Text, b.Text)
}

// compareNaturally compares text rune by rune, except that runs of digits compare by their value
func compareNaturally(a, b string) int {
	ar, br := []rune(a), []rune(b)
	i, j := 0, 0
	for i < len(ar) && j < len(br) {
		if isDigit(ar[i]) && isDigit(br[j]) {
			iEnd, jEnd := digitRunEnd(ar, i), digitRunEnd(br, j)
			aDigits := strings.TrimLeft(string(ar[i:iEnd]), "0")
			bDigits := strings.TrimLeft(string(br[j:jEnd]), "0")
			if len(aDigits) != len(bDigits) {
				if len(aDigits) < len(bDigits) {
					return -1
				}
				return 1
			}
			if c := strings.Compare(aDigits, bDigits); c != 0 {
				return c
			}
			i, j = iEnd, jEnd
			continue
		}
		if ar[i] != br[j] {
			if ar[i] < br[j] {
				return -1
			}
			return 1
		}
		i, j = i+1, j+1
	}
	switch {
	case len(ar)-i < len(br)-j:
		return -1
	case len(ar)-i > len(br)-j:
		return 1
	}
	return 0
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func digitRunEnd(r []rune, start int) int {
	end := start
	for end < len(r) && isDigit(r[end]) {
		end++
	}
	return end
}

func sortValueAt(r Row, col int) SortValue {
	if col < 0 || col >= len(r.SortValues) {
		return MissingSortValue()
	}
	return r.SortValues[col]
}

var plainDecimal = regexp.MustCompile(`^-?\d+(\.\d+)?$`)

// columnSortValues gives each row's value in column col. If every text value present is a plain
// decimal number, like a spreadsheet the column sorts by those numbers, so "-10" < "-2" < "0.25" < "0.5".
func columnSortValues(rows []Row, col int) []SortValue {
	values := make([]SortValue, len(rows))
	allDecimal := true
	for i, r := range rows {
		values[i] = sortValueAt(r, col)
		if v := values[i]; !v.Missing && v.Nums == nil && !plainDecimal.MatchString(v.Text) {
			allDecimal = false
		}
	}
	if !allDecimal {
		return values
	}
	numeric := make([]SortValue, len(values))
	for i, v := range values {
		numeric[i] = v
		if !v.Missing && v.Nums == nil {
			// a number too long for a float64 leaves the column to sort naturally
			n, err := strconv.ParseFloat(v.Text, 64)
			if err != nil {
				return values
			}
			numeric[i] = NumSortValue(n)
		}
	}
	return numeric
}

// sortRows returns a copy of rows, given in their natural order, sorted by column col (-1 keeps the
// natural order). Missing values go last in either direction, and ties keep the natural order,
// so the order stays the same across refreshes.
func sortRows(rows []Row, col int, descending bool) []Row {
	sorted := make([]Row, len(rows))
	copy(sorted, rows)
	if col < 0 {
		return sorted
	}
	values := columnSortValues(rows, col)
	order := make([]int, len(rows))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool {
		a, b := values[order[i]], values[order[j]]
		if a.Missing || b.Missing {
			return !a.Missing && b.Missing
		}
		c := compareSortValues(a, b)
		if descending {
			return c > 0
		}
		return c < 0
	})
	for i, idx := range order {
		sorted[i] = rows[idx]
	}
	return sorted
}

type sortState struct {
	// columns are the names of the sortable columns; empty if the page can't sort
	columns []string
	// column is the index in columns being sorted by, or -1 for the natural order
	column     int
	descending bool
}

func (s sortState) sortable() bool {
	return len(s.columns) > 0
}

func (s *sortState) setColumns(columns []string) {
	if !slices.Equal(columns, s.columns) {
		s.column, s.descending = -1, false
	}
	s.columns = columns
}

func (s *sortState) cycle() {
	s.column++
	if s.column >= len(s.columns) {
		s.column = -1
	}
	s.descending = false
}

// flip reverses the sort of a column; the natural order has no direction
func (s *sortState) flip() {
	if s.column >= 0 {
		s.descending = !s.descending
	}
}

// String describes the sort for the filter line, or is empty in the natural order
func (s sortState) String() string {
	if s.column < 0 || s.column >= len(s.columns) {
		return ""
	}
	arrow := "↑"
	if s.descending {
		arrow = "↓"
	}
	return fmt.Sprintf("sort: %s %s", s.columns[s.column], arrow)
}
