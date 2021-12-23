package do

import "strings"

type CsvData [][]string

func (csv CsvData) Row(row int) []string {
	if row < len(csv) {
		return (csv)[row]
	} else {
		return []string{}
	}
}

func (csv CsvData) CellByColumnName(row int, name string) string {
	first := csv.Row(0)
	var colIndex = -1
	for i := 0; i < len(first); i++ {
		if strings.TrimSpace(first[i]) == name {
			colIndex = i
		}
	}

	if colIndex == -1 {
		return ""
	}

	rowData := csv.Row(row)
	if colIndex < len(rowData) {
		return rowData[colIndex]
	}

	return ""
}

func (csv CsvData) CellByColumnPrefix(row int, prefix string) string {
	first := csv.Row(0)
	var colIndex = -1

	for i := 0; i < len(first); i++ {

		split := strings.FieldsFunc(strings.TrimSpace(first[i]), func(r rune) bool {
			return r == ';' || r == '('
		})
		if strings.TrimSpace(split[0]) == prefix {
			colIndex = i
			break
		}
	}

	if colIndex == -1 {
		return ""
	}

	rowData := csv.Row(row)
	if colIndex < len(rowData) {
		return rowData[colIndex]
	}

	return ""
}
