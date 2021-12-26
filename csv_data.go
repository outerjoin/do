package do

import (
	"fmt"
	"reflect"
	"strings"
)

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

func (csv CsvData) RowToMap(row int, modelOrType interface{}) (map[string]interface{}, []ErrorPlus) {

	output := map[string]interface{}{}
	outErrors := []ErrorPlus{}

	csvToJson := func(fld reflect.StructField, data Map, keys ...string) []ErrorPlus {
		fname := keys[len(keys)-1]
		csvColumn := fld.Tag.Get("csv")

		if csvColumn != "" {
			val := csv.CellByColumnPrefix(row, csvColumn)
			if val != "" {
				v, err := ParseType(val, fld.Type)
				if err == nil {
					data[fname] = v
				} else {
					outErrors = append(outErrors, ErrorPlus{
						Message: fmt.Sprintf("Could not parse cell value: %s", val),
						Source:  fmt.Sprintf("Row%d:%s", row, csvColumn),
					})
				}
			}
		}

		return nil
	}

	StructWalk(modelOrType, WalkConfig{"json"}, output, csvToJson)

	return output, outErrors
}
