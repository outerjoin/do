package do

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCsvData(t *testing.T) {

	var data = [][]string{
		{"harry", "merry", "terry", "larry; king"},
		{"H", "M", "T", "L"},
	}

	csvData := CsvData(data)
	assert.Equal(t, "harry", csvData.Row(0)[0])
	assert.Equal(t, "M", csvData.CellByColumnName(1, "merry"))
	assert.Equal(t, "", csvData.CellByColumnName(1, "not matching"))
	assert.Equal(t, "L", csvData.CellByColumnPrefix(1, "larry"))
}
