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
	All, Filtered []Row
}
