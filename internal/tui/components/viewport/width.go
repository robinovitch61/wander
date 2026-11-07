package viewport

import (
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
	"github.com/muesli/reflow/ansi"
)

// The viewport measures and cuts lines in terminal cells. A non-ASCII rune takes the cells lipgloss gives
// it through go-runewidth: wide characters such as CJK and most emoji take 2, combining marks and other
// zero-width runes take none and stay with the character before them, and in East Asian locales the
// ambiguous ones, such as arrows, take 2. ANSI escape sequences take none and are kept whole. Every
// ASCII byte takes one cell, tab and other control characters included, where lipgloss counts them as
// none: that is how the viewport always measured them, which keeps ASCII lines byte for byte the same.
//
// Lines of plain ASCII, the common case, take a fast path that measures and slices bytes, exactly as the
// viewport did before it measured cells.

// isPlainASCII is true when every byte of s is ASCII and none starts an escape sequence, so that each
// byte is one cell
func isPlainASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf || s[i] == ansi.Marker {
			return false
		}
	}
	return true
}

func runeWidth(r rune) int {
	if r < utf8.RuneSelf {
		return 1
	}
	return runewidth.RuneWidth(r)
}

// StringWidth is the number of terminal cells s takes in the viewport
func StringWidth(s string) int {
	return stringWidth(s)
}

func stringWidth(s string) int {
	if isPlainASCII(s) {
		return len(s)
	}
	var width int
	var inEscape bool
	for _, r := range s {
		switch {
		case r == ansi.Marker:
			inEscape = true
		case inEscape:
			inEscape = !ansi.IsTerminator(r)
		default:
			width += runeWidth(r)
		}
	}
	return width
}

// cellSlice returns the part of s that covers cells start to end, end excluded: s[start:end] for plain
// ASCII. A wide character cut by either edge is replaced by spaces for its cells inside the range, so
// the result keeps its width and is never a partial character. Zero-width runes go with the character
// before them, and escape sequences are all kept, so styling carries across the cut. Invalid UTF-8
// comes out as U+FFFD, a cell each.
func cellSlice(s string, start, end int) string {
	if end <= start {
		return ""
	}
	if isPlainASCII(s) {
		end = min(end, len(s))
		start = min(start, end)
		return s[start:end]
	}

	var b strings.Builder
	b.Grow(min(len(s), (end-start)*4))
	var col int
	var inEscape bool
	// past the range only escape sequences are kept, so the scan stops after the last one
	lastEscape := strings.LastIndexByte(s, ansi.Marker)
	// a zero-width rune at the very start belongs to the range only if the range starts there
	lastKept := start == 0
	for i, r := range s {
		if r == ansi.Marker {
			inEscape = true
			b.WriteRune(r)
			continue
		}
		if inEscape {
			inEscape = !ansi.IsTerminator(r)
			b.WriteRune(r)
			continue
		}
		w := runeWidth(r)
		if w == 0 {
			if lastKept {
				b.WriteRune(r)
			}
			continue
		}
		if col >= end {
			if i > lastEscape {
				break
			}
			lastKept = false
			continue
		}
		switch {
		case col >= start && col+w <= end:
			b.WriteRune(r)
			lastKept = true
		case col+w > start:
			// cut by an edge: blank the cells inside the range
			b.WriteString(strings.Repeat(" ", min(col+w, end)-max(col, start)))
			lastKept = false
		default:
			lastKept = false
		}
		col += w
	}
	return b.String()
}

// splitLineIntoSizedChunks cuts a line into lines of at most chunkSize cells. A wide character that would
// cross the end of a chunk starts the next one instead. Escape sequences go with the character after
// them, so a style that starts a chunk is in it, and trailing ones with the last chunk.
func splitLineIntoSizedChunks(line string, chunkSize int) []string {
	if chunkSize <= 0 {
		return []string{line}
	}
	var wrappedLines []string
	if isPlainASCII(line) {
		for len(line) > 0 {
			width := min(chunkSize, len(line))
			wrappedLines = append(wrappedLines, line[0:width])
			line = line[width:]
		}
		return wrappedLines
	}

	// chunk is the line being filled; escapes holds the escape sequences read since its last character
	var chunk, escapes strings.Builder
	var chunkWidth int
	var inEscape bool
	for _, r := range line {
		if r == ansi.Marker || inEscape {
			inEscape = r == ansi.Marker || !ansi.IsTerminator(r)
			escapes.WriteRune(r)
			continue
		}
		w := runeWidth(r)
		if w > 0 && chunkWidth > 0 && chunkWidth+w > chunkSize {
			wrappedLines = append(wrappedLines, chunk.String())
			chunk.Reset()
			chunkWidth = 0
		}
		chunk.WriteString(escapes.String())
		escapes.Reset()
		chunk.WriteRune(r)
		chunkWidth += w
	}
	chunk.WriteString(escapes.String())
	if chunk.Len() > 0 {
		if chunkWidth > 0 || len(wrappedLines) == 0 {
			wrappedLines = append(wrappedLines, chunk.String())
		} else {
			// trailing escape sequences go with the last chunk
			wrappedLines[len(wrappedLines)-1] += chunk.String()
		}
	}
	return wrappedLines
}
