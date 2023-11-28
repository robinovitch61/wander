package page

type Row struct {
	Key, Row string
}

func (r Row) String() string {
	return r.Row
}

func rowsToStrings(rows []Row) []string {
	var strs []string
	for _, row := range rows {
		strs = append(strs, row.String())
	}
	return strs
}

type data struct {
	AllRows []Row

	// FilteredRows are only the rows in AllRows that match the current filter
	FilteredRows []Row

	// these are only used in "filter but show context" mode
	FilteredSelectionNum      int
	FilteredContentIdxs       []int
	CurrentFilteredContentIdx int
}

func (d *data) IncrementFilteredSelectionNum() {
	if len(d.FilteredContentIdxs) == 0 {
		return
	}
	d.FilteredSelectionNum++
	if d.FilteredSelectionNum >= len(d.FilteredContentIdxs) {
		d.FilteredSelectionNum = 0
	}
	d.CurrentFilteredContentIdx = d.FilteredContentIdxs[d.FilteredSelectionNum]
}

func (d *data) DecrementFilteredSelectionNum() {
	if len(d.FilteredContentIdxs) == 0 {
		return
	}
	d.FilteredSelectionNum--
	if d.FilteredSelectionNum < 0 {
		d.FilteredSelectionNum = len(d.FilteredContentIdxs) - 1
	}
	d.CurrentFilteredContentIdx = d.FilteredContentIdxs[d.FilteredSelectionNum]
}
