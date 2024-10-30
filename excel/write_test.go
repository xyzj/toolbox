package excel

import (
	_ "embed"
	"testing"
)

//go:embed sz_daily_record.xlsx
var szDailyRecordFile []byte

func TestExcel(t *testing.T) {
	fd, err := NewExcelFromBinary(szDailyRecordFile, "daily")
	if err != nil {
		t.Fatal(err)
		return
	}
	cell1, _ := fd.GetCell("Sheet1", 2, 0)
	println(cell1.String())
	fd.Save()
}
