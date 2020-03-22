package parser

import (
	"encoding/csv"
	"io"
)

func ParseCSV(reader io.Reader, fpr int) ([][]string, error) {
	r := csv.NewReader(reader)
	r.FieldsPerRecord = fpr
	return r.ReadAll()
}
