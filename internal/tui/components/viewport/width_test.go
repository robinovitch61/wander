package viewport

import (
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
	"github.com/robinovitch61/wander/internal/tui/constants"
)

// TestMain measures as outside East Asian locales whatever the locale the tests run in, since
// go-runewidth counts ambiguous characters such as ↑ as two cells under LANG=ja_JP and the like
func TestMain(m *testing.M) {
	runewidth.DefaultCondition.EastAsianWidth = false
	os.Exit(m.Run())
}

func TestStringWidth(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"plain text", 10},
		{"tab\there", 8},
		{"日本語", 6},
		{"a😀b", 4},
		{"e\u0301te\u0301", 3}, // é as e + combining acute
		{"Submitted ↓", 11},
		{"\x1b[31mred\x1b[0m", 3},
		{"\x1b[1m日本\x1b[0m", 4},
	}
	for _, c := range cases {
		if got := stringWidth(c.in); got != c.want {
			t.Errorf("stringWidth(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestCellSlice(t *testing.T) {
	cases := []struct {
		in         string
		start, end int
		want       string
	}{
		{"abcdef", 1, 4, "bcd"},
		{"abcdef", 4, 10, "ef"},
		{"abcdef", 3, 3, ""},
		{"abcdef", 4, 2, ""},
		// wide characters take two cells each
		{"日本語", 0, 4, "日本"},
		{"日本語", 2, 6, "本語"},
		// a cut inside a wide character blanks its cells inside the range, keeping the width
		{"日本語", 1, 4, " 本"},
		{"日本語", 0, 3, "日 "},
		{"日本語", 1, 2, " "},
		{"a😀b", 2, 4, " b"},
		{"a😀b", 0, 2, "a "},
		// combining marks stay with their base character
		{"e\u0301te\u0301", 0, 1, "e\u0301"},
		{"e\u0301te\u0301", 1, 3, "te\u0301"},
		{"e\u0301te\u0301", 1, 2, "t"},
		// a combining mark at the start belongs to the range only if it starts there
		{"\u0301ab", 0, 1, "\u0301a"},
		{"\u0301ab", 1, 2, "b"},
		// escape sequences take no cells and are kept, so styling carries across the cut
		{"\x1b[31mred\x1b[0m", 1, 3, "\x1b[31med\x1b[0m"},
		{"\x1b[31mred\x1b[0mblue", 2, 5, "\x1b[31md\x1b[0mbl"},
		{"\x1b[1m日本\x1b[0m", 1, 4, "\x1b[1m 本\x1b[0m"},
		{"Submitted ↓  Status", 8, 13, "d ↓  "},
		{"Submitted ↓  Status", 10, 11, "↓"},
	}
	for _, c := range cases {
		got := cellSlice(c.in, c.start, c.end)
		if got != c.want {
			t.Errorf("cellSlice(%q, %d, %d) = %q, want %q", c.in, c.start, c.end, got, c.want)
		}
		if c.end > c.start && !strings.Contains(c.in, "\u0301") {
			if w, want := stringWidth(got), min(c.end, stringWidth(c.in))-min(c.start, stringWidth(c.in)); w != want {
				t.Errorf("cellSlice(%q, %d, %d) is %d cells, want %d", c.in, c.start, c.end, w, want)
			}
		}
	}
}

func TestEastAsianWidths(t *testing.T) {
	runewidth.DefaultCondition.EastAsianWidth = true
	defer func() { runewidth.DefaultCondition.EastAsianWidth = false }()
	// ambiguous characters take two cells in East Asian locales; wide ones take two anyway
	if got := stringWidth("Submitted ↓"); got != 12 {
		t.Errorf("stringWidth(\"Submitted ↓\") = %d, want 12", got)
	}
	if got := cellSlice("Count ↑  x", 6, 8); got != "↑" {
		t.Errorf("cellSlice = %q, want ↑", got)
	}
}

// invalid UTF-8 comes out as U+FFFD, as ranging over the string reads it
func TestCellSliceInvalidUTF8(t *testing.T) {
	if got, want := cellSlice("日\xffab", 0, 10), "日\uFFFDab"; got != want {
		t.Errorf("cellSlice = %q, want %q", got, want)
	}
}

func TestSplitLineIntoSizedChunks(t *testing.T) {
	cases := []struct {
		in   string
		size int
		want []string
	}{
		{"abcdefg", 3, []string{"abc", "def", "g"}},
		{"日本語です", 4, []string{"日本", "語で", "す"}},
		// a wide character that would cross the edge starts the next chunk
		{"a日本語", 4, []string{"a日", "本語"}},
		{"a日b", 2, []string{"a", "日", "b"}},
		{"e\u0301e\u0301e\u0301", 2, []string{"e\u0301e\u0301", "e\u0301"}},
		{"\x1b[31mabcd\x1b[0m", 2, []string{"\x1b[31mab", "cd\x1b[0m"}},
		// an escape before a character goes with it, so a style opened at a chunk edge styles the next chunk
		{"abc\x1b[31mdef\x1b[0m", 3, []string{"abc", "\x1b[31mdef\x1b[0m"}},
		{"ab\x1b[1m\x1b[31m日\x1b[0m", 3, []string{"ab", "\x1b[1m\x1b[31m日\x1b[0m"}},
	}
	for _, c := range cases {
		if got := splitLineIntoSizedChunks(c.in, c.size); !reflect.DeepEqual(got, c.want) {
			t.Errorf("splitLineIntoSizedChunks(%q, %d) = %q, want %q", c.in, c.size, got, c.want)
		}
	}
}

func TestVisiblePartOfLineWide(t *testing.T) {
	m := New(6, 10, false)
	line := "日本語のテキスト" // 16 cells
	m.SetContent([]string{line})
	for x := 0; x <= stringWidth(line)-m.width; x++ {
		m.xOffset = x
		got := m.getVisiblePartOfLine(line)
		if w := stringWidth(got); w != m.width {
			t.Errorf("xOffset %d: %q is %d cells, want %d", x, got, w, m.width)
		}
	}
	cases := []struct {
		xOffset int
		want    string
	}{
		// 日本語 fills the six cells; the "..." replaces the last three, blanking the half of 語 left over
		{0, "日 ..."},
		// キスト after the "...", which covers キ and half of ス
		{10, "... ト"},
		// cells 3-8 are the right half of 本, 語, の and the left half of テ; "..." covers both ends
		{3, "......"},
	}
	for _, c := range cases {
		m.xOffset = c.xOffset
		if got := m.getVisiblePartOfLine(line); got != c.want {
			t.Errorf("xOffset %d: got %q, want %q", c.xOffset, got, c.want)
		}
	}
}

func TestCompactTableWide(t *testing.T) {
	sep := constants.TableSeparator
	header := []string{"Job    " + sep + "Status " + sep}
	content := []string{
		"日本語 " + sep + "running" + sep,
		"web    " + sep + "dead   " + sep,
	}
	gotHeader, gotContent := horizontallyCompact(header, content)
	pad := constants.TablePadding
	wantHeader := []string{"Job   " + pad + "Status " + pad}
	wantContent := []string{"日本語" + pad + "running" + pad, "web   " + pad + "dead   " + pad}
	if !reflect.DeepEqual(gotHeader, wantHeader) || !reflect.DeepEqual(gotContent, wantContent) {
		t.Errorf("got %q %q, want %q %q", gotHeader, gotContent, wantHeader, wantContent)
	}
}

// The functions below are the viewport's byte-based measuring and slicing before it counted cells,
// kept to check that plain ASCII lines come out byte for byte the same.

func oldSplitLineIntoSizedChunks(line string, chunkSize int) []string {
	var wrappedLines []string
	for {
		lineWidth := len(line)
		if lineWidth == 0 {
			break
		}
		width := chunkSize
		if lineWidth < chunkSize {
			width = lineWidth
		}
		wrappedLines = append(wrappedLines, line[0:width])
		line = line[width:]
	}
	return wrappedLines
}

func oldGetVisiblePartOfLine(line string, xOffset, width int) string {
	rightTrimmedLineLength := len(strings.TrimRight(line, " "))
	end := min(len(line), xOffset+width)
	start := min(end, xOffset)
	line = line[start:end]
	if xOffset+width < rightTrimmedLineLength {
		truncate := max(0, len(line)-3)
		line = line[:truncate] + "..."
	}
	if xOffset > 0 {
		line = "..." + line[min(len(line), 3):]
	}
	return line
}

func oldGetWrappedLines(line string, width int) []string {
	if len(line) < width {
		return []string{line}
	}
	line = strings.ReplaceAll(line, constants.TableSeparator, constants.TablePadding)
	line = strings.TrimRight(line, " ")
	return oldSplitLineIntoSizedChunks(line, width)
}

func oldHorizontallyCompact(header, content []string) ([]string, []string) {
	allLines := append(append([]string{}, header...), content...)
	numCols := len(strings.Split(allLines[0], constants.TableSeparator))
	maxLen := make([]int, numCols)
	for _, line := range allLines {
		columns := strings.Split(line, constants.TableSeparator)
		if len(columns) != numCols {
			return header, content
		}
		for i, column := range columns {
			maxLen[i] = max(maxLen[i], len(strings.TrimRight(column, " ")))
		}
	}
	compact := func(lines []string) []string {
		var output []string
		for _, line := range lines {
			columns := strings.Split(line, constants.TableSeparator)
			for i, column := range columns {
				columns[i] = strings.TrimRight(column, " ") + strings.Repeat(" ", maxLen[i]-len(strings.TrimRight(column, " ")))
			}
			output = append(output, strings.Join(columns, constants.TablePadding))
		}
		return output
	}
	return compact(header), compact(content)
}

func randomASCIILine(r *rand.Rand, maxLen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_.:/{}\"   \t"
	b := make([]byte, r.Intn(maxLen+1))
	for i := range b {
		b[i] = chars[r.Intn(len(chars))]
	}
	return string(b)
}

func randomASCIITable(r *rand.Rand) []string {
	numCols, numRows := 1+r.Intn(5), 1+r.Intn(6)
	widths := make([]int, numCols)
	for i := range widths {
		widths[i] = 1 + r.Intn(15)
	}
	var lines []string
	for row := 0; row < numRows; row++ {
		var line string
		for _, w := range widths {
			cell := randomASCIILine(r, w)
			cell = strings.ReplaceAll(cell, "\t", " ")
			line += cell + strings.Repeat(" ", w-len(cell)) + constants.TableSeparator
		}
		lines = append(lines, line)
	}
	return lines
}

func TestASCIIUnchanged(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 20000; i++ {
		line := randomASCIILine(r, 120)
		width := 1 + r.Intn(100)
		xOffset := r.Intn(130)
		if got, want := splitLineIntoSizedChunks(line, width), oldSplitLineIntoSizedChunks(line, width); !reflect.DeepEqual(got, want) {
			t.Fatalf("splitLineIntoSizedChunks(%q, %d) = %q, before %q", line, width, got, want)
		}
		m := Model{width: width, xOffset: xOffset}
		if got, want := m.getVisiblePartOfLine(line), oldGetVisiblePartOfLine(line, xOffset, width); got != want {
			t.Fatalf("getVisiblePartOfLine(%q) at x %d width %d = %q, before %q", line, xOffset, width, got, want)
		}
		if got, want := m.getWrappedLines(line), oldGetWrappedLines(line, width); !reflect.DeepEqual(got, want) {
			t.Fatalf("getWrappedLines(%q, %d) = %q, before %q", line, width, got, want)
		}

		table := randomASCIITable(r)
		if got, want := m.getWrappedLines(table[0]), oldGetWrappedLines(table[0], width); !reflect.DeepEqual(got, want) {
			t.Fatalf("getWrappedLines(%q, %d) = %q, before %q", table[0], width, got, want)
		}
		gotHeader, gotContent := horizontallyCompact(append([]string{}, table[:1]...), append([]string{}, table[1:]...))
		wantHeader, wantContent := oldHorizontallyCompact(table[:1], table[1:])
		if !reflect.DeepEqual(gotHeader, wantHeader) || !reflect.DeepEqual(gotContent, wantContent) {
			t.Fatalf("horizontallyCompact(%q) = %q %q, before %q %q", table, gotHeader, gotContent, wantHeader, wantContent)
		}
	}
}

func BenchmarkVisiblePartOfLineASCII(b *testing.B) {
	line := strings.Repeat("2026-11-06T15:04:05Z some log line with ascii text ", 8)
	m := Model{width: 100, xOffset: 25}
	for i := 0; i < b.N; i++ {
		m.getVisiblePartOfLine(line)
	}
}

func BenchmarkVisiblePartOfLineWide(b *testing.B) {
	line := strings.Repeat("2026-11-06T15:04:05Z 日本語のログ行 with some text ", 8)
	m := Model{width: 100, xOffset: 25}
	for i := 0; i < b.N; i++ {
		m.getVisiblePartOfLine(line)
	}
}

func BenchmarkVisiblePartOfLineASCIIBefore(b *testing.B) {
	line := strings.Repeat("2026-11-06T15:04:05Z some log line with ascii text ", 8)
	for i := 0; i < b.N; i++ {
		oldGetVisiblePartOfLine(line, 25, 100)
	}
}

// an escape after the range used to cost a scan of the rest of the line for each rune past the range
func BenchmarkCellSliceStyledTail(b *testing.B) {
	for _, n := range []int{5000, 10000, 20000, 40000} {
		line := "\x1b[31m" + strings.Repeat("é", n) + "\x1b[0m"
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				cellSlice(line, 0, 100)
			}
		})
	}
}
