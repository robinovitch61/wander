package page

import (
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func keysOf(rows []Row) []string {
	var keys []string
	for _, r := range rows {
		keys = append(keys, r.Key)
	}
	return keys
}

func rowsWith(values ...SortValue) []Row {
	var rows []Row
	for i, v := range values {
		rows = append(rows, Row{Key: string(rune('a' + i)), SortValues: []SortValue{v}})
	}
	return rows
}

func TestTextSortValue(t *testing.T) {
	cases := []struct {
		in   string
		want SortValue
	}{
		{"Web", SortValue{Text: "web"}},
		{" 42 ", SortValue{Text: "42"}},
		{"-", MissingSortValue()},
		{"", MissingSortValue()},
	}
	for _, c := range cases {
		if got := TextSortValue(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("TextSortValue(%q) = %+v, want %+v", c.in, got, c.want)
		}
	}
}

func TestCompareSortValues(t *testing.T) {
	cases := []struct {
		a, b SortValue
		want int
	}{
		{NumSortValue(2), NumSortValue(10), -1},
		{NumSortValue(10), NumSortValue(2), 1},
		{NumSortValue(3), NumSortValue(3), 0},
		// later numbers break ties of earlier ones
		{NumSortValue(1, 5), NumSortValue(1, 3), 1},
		{NumSortValue(1, 5), NumSortValue(2, 0), -1},
		// runs of digits in text compare by value
		{TextSortValue("10"), TextSortValue("2"), 1},
		{TextSortValue("web-9"), TextSortValue("web-10"), -1},
		{TextSortValue("007"), TextSortValue("7"), 0},
		{TextSortValue("apple"), TextSortValue("Banana"), -1},
		{TextSortValue("ABC"), TextSortValue("abc"), 0},
		{NumSortValue(99), TextSortValue("a"), -1},
		{TextSortValue("a"), NumSortValue(99), 1},
	}
	for _, c := range cases {
		if got := compareSortValues(c.a, c.b); got != c.want {
			t.Errorf("compareSortValues(%+v, %+v) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestSortRowsNumbers(t *testing.T) {
	// times sort as numbers (nanoseconds), not as their rendered text
	rows := rowsWith(NumSortValue(1700000000e9), NumSortValue(900e9), NumSortValue(1800000000e9))
	if got, want := keysOf(sortRows(rows, 0, false)), []string{"b", "a", "c"}; !reflect.DeepEqual(got, want) {
		t.Errorf("ascending = %v, want %v", got, want)
	}
	if got, want := keysOf(sortRows(rows, 0, true)), []string{"c", "a", "b"}; !reflect.DeepEqual(got, want) {
		t.Errorf("descending = %v, want %v", got, want)
	}
}

func TestSortRowsText(t *testing.T) {
	rows := rowsWith(TextSortValue("zeta"), TextSortValue("Alpha"), TextSortValue("beta"))
	if got, want := keysOf(sortRows(rows, 0, false)), []string{"b", "c", "a"}; !reflect.DeepEqual(got, want) {
		t.Errorf("ascending = %v, want %v", got, want)
	}
	if got, want := keysOf(sortRows(rows, 0, true)), []string{"a", "c", "b"}; !reflect.DeepEqual(got, want) {
		t.Errorf("descending = %v, want %v", got, want)
	}
}

func TestSortRowsTiesKeepNaturalOrder(t *testing.T) {
	rows := rowsWith(TextSortValue("x"), TextSortValue("y"), TextSortValue("x"), TextSortValue("y"))
	if got, want := keysOf(sortRows(rows, 0, false)), []string{"a", "c", "b", "d"}; !reflect.DeepEqual(got, want) {
		t.Errorf("ascending = %v, want %v", got, want)
	}
	// ties stay in natural order in the other direction too
	if got, want := keysOf(sortRows(rows, 0, true)), []string{"b", "d", "a", "c"}; !reflect.DeepEqual(got, want) {
		t.Errorf("descending = %v, want %v", got, want)
	}
}

func TestSortRowsMissingLast(t *testing.T) {
	rows := rowsWith(MissingSortValue(), NumSortValue(2), MissingSortValue(), NumSortValue(1))
	if got, want := keysOf(sortRows(rows, 0, false)), []string{"d", "b", "a", "c"}; !reflect.DeepEqual(got, want) {
		t.Errorf("ascending = %v, want %v", got, want)
	}
	if got, want := keysOf(sortRows(rows, 0, true)), []string{"b", "d", "a", "c"}; !reflect.DeepEqual(got, want) {
		t.Errorf("descending = %v, want %v", got, want)
	}
}

func TestSortRowsNatural(t *testing.T) {
	rows := rowsWith(NumSortValue(3), NumSortValue(1), NumSortValue(2))
	for _, descending := range []bool{false, true} {
		if got, want := keysOf(sortRows(rows, -1, descending)), []string{"a", "b", "c"}; !reflect.DeepEqual(got, want) {
			t.Errorf("natural (descending %v) = %v, want %v", descending, got, want)
		}
	}
	sortRows(rows, 0, false)
	if got := keysOf(rows); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Errorf("sortRows changed its input: %v", got)
	}
}

func TestSortStateCycleAndFlip(t *testing.T) {
	s := sortState{column: -1}
	s.setColumns([]string{"Job", "Submitted"})

	// the natural order has no direction to flip
	s.flip()
	if s.column != -1 || s.descending {
		t.Errorf("flip in natural order = %+v", s)
	}

	var seen []int
	for i := 0; i < 4; i++ {
		s.cycle()
		seen = append(seen, s.column)
	}
	if want := []int{0, 1, -1, 0}; !reflect.DeepEqual(seen, want) {
		t.Errorf("cycle = %v, want %v", seen, want)
	}

	s.flip()
	if !s.descending {
		t.Errorf("flip on a column should reverse it")
	}
	// a refresh with the same columns keeps the sort
	s.setColumns([]string{"Job", "Submitted"})
	if s.column != 0 || !s.descending {
		t.Errorf("after refresh = %+v", s)
	}
	// moving to the next column starts ascending
	s.cycle()
	if s.column != 1 || s.descending {
		t.Errorf("after cycle = %+v", s)
	}
	// different columns reset it
	s.setColumns([]string{"Key", "Value"})
	if s.column != -1 || s.descending {
		t.Errorf("after new columns = %+v", s)
	}
}

func TestSortStateString(t *testing.T) {
	s := sortState{column: -1}
	s.setColumns([]string{"Job", "Since Submit"})
	if got := s.String(); got != "" {
		t.Errorf("natural order = %q, want no label", got)
	}
	s.cycle()
	if got, want := s.String(), "sort: Job ↑"; got != want {
		t.Errorf("ascending = %q, want %q", got, want)
	}
	s.cycle()
	s.flip()
	if got, want := s.String(), "sort: Since Submit ↓"; got != want {
		t.Errorf("descending = %q, want %q", got, want)
	}
}

func keyPress(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func newSortedPage(ctx bool, rows []Row) Model {
	m := New(Config{Width: 100, Height: 30, SelectionEnabled: true}, false, false, ctx)
	m.SetHeader([]string{"Job Status"})
	m.SetSortColumns([]string{"Job", "Status"})
	m.SetAllPageRows(rows)
	m.SetLoading(false)
	return m
}

func statusRows(statuses ...string) []Row {
	var rows []Row
	for i, s := range statuses {
		k := string(rune('a' + i))
		rows = append(rows, Row{Key: k, Row: k + " " + s, SortValues: []SortValue{TextSortValue(k), TextSortValue(s)}})
	}
	return rows
}

func selectedKey(t *testing.T, m Model) string {
	r, err := m.GetSelectedPageRow()
	if err != nil {
		t.Fatal(err)
	}
	return r.Key
}

// a refresh can change a sorted column's values and so the order; the cursor stays on its row
func TestRefreshKeepsCursorOnRowWhenSorted(t *testing.T) {
	m := newSortedPage(false, statusRows("pending", "pending", "pending", "pending"))
	m, _ = m.Update(keyPress("S"))
	m, _ = m.Update(keyPress("S")) // sort: Status asc
	m, _ = m.Update(keyPress("j")) // cursor on b
	if got := selectedKey(t, m); got != "b" {
		t.Fatalf("setup: selected %q", got)
	}
	// periodic refresh: job a started running; same keys, same natural order
	m.SetAllPageRows(statusRows("running", "pending", "pending", "pending"))
	if got := selectedKey(t, m); got != "b" {
		t.Errorf("after refresh the cursor is on %q, was on b", got)
	}
}

// words that parse as floats, like "nan" and "Infinity", are text
func TestTextSortValueWordsAreText(t *testing.T) {
	rows := rowsWith(TextSortValue("3"), TextSortValue("nan"), TextSortValue("1"), TextSortValue("2"))
	got := keysOf(sortRows(rows, 0, false))
	// numbers ascending (c=1, d=2, a=3) in some position relative to "nan"
	var nums []string
	for _, k := range got {
		if k != "b" {
			nums = append(nums, k)
		}
	}
	if want := []string{"c", "d", "a"}; !reflect.DeepEqual(nums, want) {
		t.Errorf("ascending = %v: numbers out of order %v, want %v", got, nums, want)
	}
	if v := TextSortValue("Infinity"); v.Nums != nil {
		t.Errorf("TextSortValue(\"Infinity\") = %+v, a word read as a number", v)
	}
}

// version-like values sort by their numbers
func TestSortRowsVersionLikeText(t *testing.T) {
	rows := rowsWith(TextSortValue("1.10.1"), TextSortValue("2"), TextSortValue("1.10"), TextSortValue("1.9"))
	got := keysOf(sortRows(rows, 0, false))
	if want := []string{"d", "c", "a", "b"}; !reflect.DeepEqual(got, want) {
		t.Errorf("ascending = %v, want %v (1.9, 1.10, 1.10.1, 2)", got, want)
	}
}

// with a context filter, n after a sort goes to the next match below the cursor in display order
func TestContextFilterAfterSort(t *testing.T) {
	m := newSortedPage(true, statusRows("x", "y", "x", "y", "x"))
	for _, k := range []string{"/", "x"} {
		m, _ = m.Update(keyPress(k))
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = m.Update(keyPress("S"))
	m, _ = m.Update(keyPress("S")) // Status asc: a c e (x) b d (y)
	m, _ = m.Update(keyPress("D")) // Status desc: b d a c e
	if got := keysOf(m.pageData.FilteredRows); !reflect.DeepEqual(got, []string{"b", "d", "a", "c", "e"}) {
		t.Fatalf("order %v", got)
	}
	if got := selectedKey(t, m); got != "a" {
		t.Errorf("cursor on %q, want a", got)
	}
	m, _ = m.Update(keyPress("n"))
	if got := selectedKey(t, m); got != "c" {
		t.Errorf("after n cursor on %q, want c", got)
	}
}

// with a plain filter applied, S keeps the cursor on its row
func TestPlainFilterAfterSort(t *testing.T) {
	m := newSortedPage(false, statusRows("x", "y", "x", "y", "x"))
	for _, k := range []string{"/", "x"} {
		m, _ = m.Update(keyPress(k))
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = m.Update(keyPress("j"))
	m, _ = m.Update(keyPress("S"))
	m, _ = m.Update(keyPress("D"))
	if got := selectedKey(t, m); got != "c" {
		t.Errorf("cursor on %q, want c", got)
	}
	if got := keysOf(m.pageData.FilteredRows); !reflect.DeepEqual(got, []string{"e", "c", "a"}) {
		t.Errorf("order %v", got)
	}
}

// S and D typed into the filter stay text
func TestSDTypeableInFilter(t *testing.T) {
	m := newSortedPage(false, statusRows("x", "y"))
	for _, k := range []string{"/", "S", "D"} {
		m, _ = m.Update(keyPress(k))
	}
	if m.filter.Value() != "SD" || m.sort.column != -1 {
		t.Errorf("filter %q sort %+v", m.filter.Value(), m.sort)
	}
}

// app.go's "No results" message rows have no sort values, so S does nothing and no label shows
func TestMessageRowsDoNotSort(t *testing.T) {
	m := New(Config{Width: 100, Height: 30, SelectionEnabled: true}, false, false, false)
	m.SetHeader([]string{"Job"})
	m.SetSortColumns([]string{"Job", "Type"})
	m.SetAllPageRows([]Row{})
	m.SetLoading(false)
	m.SetHeader([]string{"Error"})
	m.SetAllPageRows([]Row{{Row: "No results. Is the cluster empty or was no nomad token provided?"}, {Row: "Press q or ctrl+c to quit."}})
	m.SetViewportSelectionEnabled(false)
	m, _ = m.Update(keyPress("S"))
	if s := m.sort.String(); s != "" {
		t.Errorf("error page shows %q", s)
	}
}

// a sort that was on when the rows became message rows shows no label
func TestMessageRowsHideSort(t *testing.T) {
	m := newSortedPage(false, statusRows("x", "y"))
	m, _ = m.Update(keyPress("S"))
	m.SetAllPageRows([]Row{{Row: "No results. Is the cluster empty or was no nomad token provided?"}})
	if m.sortable() {
		t.Errorf("message rows are sortable")
	}
	if strings.Contains(m.View(), "sort:") {
		t.Errorf("message rows show a sort label")
	}
}

// in the natural order a refresh keeps the cursor's index, as wander always has
func TestRefreshKeepsCursorIndexInNaturalOrder(t *testing.T) {
	m := newSortedPage(false, statusRows("x", "y", "z"))
	m, _ = m.Update(keyPress("j"))
	m.SetAllPageRows(statusRows("x", "y", "z")[1:])
	if got := selectedKey(t, m); got != "c" {
		t.Errorf("after refresh the cursor is on %q, want c at the same index", got)
	}
}

// task keys change between refreshes; their ID keeps the cursor on the same task
func TestRefreshKeepsCursorByRowID(t *testing.T) {
	withKeys := func(suffix string, statuses ...string) []Row {
		rows := statusRows(statuses...)
		for i := range rows {
			rows[i].ID = rows[i].Key
			rows[i].Key += suffix
		}
		return rows
	}
	m := newSortedPage(false, withKeys("-v1", "pending", "pending", "pending"))
	m, _ = m.Update(keyPress("S"))
	m, _ = m.Update(keyPress("S"))
	m, _ = m.Update(keyPress("j"))
	m.SetAllPageRows(withKeys("-v2", "running", "pending", "pending"))
	if got := selectedKey(t, m); got != "b-v2" {
		t.Errorf("after refresh the cursor is on %q, want b-v2", got)
	}
}

// with a context filter, a refresh that moves the selected match keeps it the current match
func TestRefreshKeepsContextFilterMatch(t *testing.T) {
	m := newSortedPage(true, statusRows("x", "y", "x", "y"))
	for _, k := range []string{"/", "x"} {
		m, _ = m.Update(keyPress(k))
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = m.Update(keyPress("S"))
	m, _ = m.Update(keyPress("S")) // Status asc: a c (x) b d (y), cursor on a, match 1/2
	m, _ = m.Update(keyPress("n")) // cursor on c, match 2/2
	// d becomes "w" and moves to the top, moving c down one
	m.SetAllPageRows(statusRows("x", "y", "x", "w"))
	if got := selectedKey(t, m); got != "c" {
		t.Fatalf("after refresh the cursor is on %q, want c", got)
	}
	if got := m.pageData.FilteredRows[m.pageData.CurrentFilteredContentIdx].Key; got != "c" {
		t.Errorf("current match is %q, want c", got)
	}
}

// a column of plain decimal numbers in text, negatives included, sorts by value
func TestSortRowsDecimalsAndNegatives(t *testing.T) {
	rows := rowsWith(TextSortValue("0.5"), TextSortValue("0.25"), TextSortValue("-2"), TextSortValue("-10"), TextSortValue("3"))
	got := keysOf(sortRows(rows, 0, false))
	if want := []string{"d", "c", "b", "a", "e"}; !reflect.DeepEqual(got, want) {
		t.Errorf("ascending = %v, want %v (-10, -2, 0.25, 0.5, 3)", got, want)
	}
}

// a numeric column with decimals and negatives sorts by value both ways
func TestSortRowsNumericTextColumn(t *testing.T) {
	rows := rowsWith(TextSortValue("3"), TextSortValue("-2"), TextSortValue("0.5"), MissingSortValue(), TextSortValue("-10"), TextSortValue("0.25"))
	if got, want := keysOf(sortRows(rows, 0, false)), []string{"e", "b", "f", "c", "a", "d"}; !reflect.DeepEqual(got, want) {
		t.Errorf("ascending = %v, want %v (-10, -2, 0.25, 0.5, 3, missing)", got, want)
	}
	if got, want := keysOf(sortRows(rows, 0, true)), []string{"a", "c", "f", "b", "e", "d"}; !reflect.DeepEqual(got, want) {
		t.Errorf("descending = %v, want %v (3, 0.5, 0.25, -2, -10, missing)", got, want)
	}
}

// one 3-part version makes the column text, so it sorts naturally
func TestSortRowsVersionColumn(t *testing.T) {
	rows := rowsWith(TextSortValue("2"), TextSortValue("1.10.1"), TextSortValue("1.9"), TextSortValue("1.10"))
	if got, want := keysOf(sortRows(rows, 0, false)), []string{"c", "d", "b", "a"}; !reflect.DeepEqual(got, want) {
		t.Errorf("ascending = %v, want %v (1.9, 1.10, 1.10.1, 2)", got, want)
	}
}

// numbers mixed with words sort naturally
func TestSortRowsMixedNumbersAndWords(t *testing.T) {
	rows := rowsWith(TextSortValue("web"), TextSortValue("10"), TextSortValue("Api"), TextSortValue("9"))
	if got, want := keysOf(sortRows(rows, 0, false)), []string{"d", "b", "c", "a"}; !reflect.DeepEqual(got, want) {
		t.Errorf("ascending = %v, want %v (9, 10, api, web)", got, want)
	}
}

// words that parse as floats keep a column from sorting by value
func TestSortRowsFloatWordsAreNotNumbers(t *testing.T) {
	rows := rowsWith(TextSortValue("2"), TextSortValue("Inf"), TextSortValue("10"), TextSortValue("nan"))
	if got, want := keysOf(sortRows(rows, 0, false)), []string{"a", "c", "b", "d"}; !reflect.DeepEqual(got, want) {
		t.Errorf("ascending = %v, want %v (2, 10, inf, nan)", got, want)
	}
}

// a number too long for a float64 leaves the whole column to sort naturally
func TestSortRowsHugeNumberSortsNaturally(t *testing.T) {
	huge := "1" + strings.Repeat("0", 400)
	rows := rowsWith(TextSortValue(huge), TextSortValue("-2"), TextSortValue("3"))
	if got, want := keysOf(sortRows(rows, 0, false)), []string{"b", "c", "a"}; !reflect.DeepEqual(got, want) {
		t.Errorf("ascending = %v, want %v", got, want)
	}
}
