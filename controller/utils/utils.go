package utils

import (
	"encoding/csv"
	"os"
)

func ReadCSVData(file string) (workload [][]string, err error) {
	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return
	}
	return records, err
}
