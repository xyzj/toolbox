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

func TestExcelJSONRoundTrip(t *testing.T) {
	fd, err := NewExcel("roundtrip")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = fd.AddSheet("SheetA"); err != nil {
		t.Fatal(err)
	}
	fd.AddRow("a1", "a2")
	fd.AddRow("b1", "b2")

	if _, err = fd.AddSheet("SheetB"); err != nil {
		t.Fatal(err)
	}
	fd.AddRow("c1")

	js, err := fd.ToJSON()
	if err != nil {
		t.Fatal(err)
	}

	nfd, err := NewExcelFromJSON([]byte(js), "from_json")
	if err != nil {
		t.Fatal(err)
	}

	rowsA := nfd.GetRows("SheetA")
	if len(rowsA) != 2 || len(rowsA[0]) != 2 || rowsA[0][0] != "a1" || rowsA[1][1] != "b2" {
		t.Fatalf("SheetA rows mismatch: %#v", rowsA)
	}

	rowsB := nfd.GetRows("SheetB")
	if len(rowsB) != 1 || len(rowsB[0]) != 1 || rowsB[0][0] != "c1" {
		t.Fatalf("SheetB rows mismatch: %#v", rowsB)
	}
}

func TestExcelToJSONWriteFile(t *testing.T) {
	fd, err := NewExcel("上海项目开发进度0707.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	js, err := fd.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	println(js)

}
