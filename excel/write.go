// Package excel excel功能模块
package excel

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"codeberg.org/tealeg/xlsx/v4"
)

var errSheetNotFound = func(sheetname string) error {
	if sheetname == "" {
		return nil
	}
	return errors.New("sheet " + sheetname + " not found")
}

// FileData Excel文件结构
type FileData struct {
	fileName   string
	colStyle   *xlsx.Style
	writeFile  *xlsx.File
	writeSheet *xlsx.Sheet
}

// GetSheet 获取Sheet,可编辑
func (fd *FileData) GetSheet(sheetname string) (*xlsx.Sheet, error) {
	sheet, ok := fd.writeFile.Sheet[sheetname]
	if !ok {
		return nil, errSheetNotFound(sheetname)
	}
	return sheet, nil
}

// GetCell 获取单元格，可进行赋值，合并等操作
func (fd *FileData) GetCell(sheetname string, row, col int) (*xlsx.Cell, error) {
	sheet, ok := fd.writeFile.Sheet[sheetname]
	if !ok {
		return nil, errSheetNotFound(sheetname)
	}
	return sheet.Cell(row, col)
}

// GetRows 获取制定sheet的所有行
func (fd *FileData) GetRows(sheetname string) [][]string {
	sheet, ok := fd.writeFile.Sheet[sheetname]
	if !ok {
		return make([][]string, 0)
	}
	ss := make([][]string, 0, sheet.MaxRow)
	l := sheet.MaxCol
	sheet.ForEachRow(func(r *xlsx.Row) error {
		rs := make([]string, 0, l)
		r.ForEachCell(func(c *xlsx.Cell) error {
			rs = append(rs, c.Value)
			return nil
		})
		if len(rs) > 0 {
			ss = append(ss, rs)
		}
		return nil
	})
	return ss
}

// AddSheet 添加sheet
// sheetname sheet名称
func (fd *FileData) AddSheet(sheetname string) (*xlsx.Sheet, error) {
	var err error
	fd.writeSheet, err = fd.writeFile.AddSheet(sheetname)
	if err != nil {
		return nil, errors.New("excel-sheet创建失败:" + err.Error())
	}
	return fd.writeSheet, nil
}

// AddRowInSheet 在指定sheet添加行
// cells： 每个单元格的数据，任意格式
func (fd *FileData) AddRowInSheet(sheetname string, cells ...interface{}) error {
	sheet, ok := fd.writeFile.Sheet[sheetname]
	if !ok {
		return fmt.Errorf("sheet " + sheetname + " not found")
	}
	row := sheet.AddRow()
	row.SetHeight(15)
	// row.WriteSlice(cells, -1)
	for _, v := range cells {
		row.AddCell().SetValue(v)
	}
	return nil
}

// AddRow 在当前sheet添加行
// cells： 每个单元格的数据，任意格式
func (fd *FileData) AddRow(cells ...interface{}) {
	if fd.writeSheet == nil {
		fd.writeSheet, _ = fd.AddSheet("newsheet" + strconv.Itoa(len(fd.writeFile.Sheets)+1))
	}
	row := fd.writeSheet.AddRow()
	row.SetHeight(15)
	row.WriteSlice(cells, -1)
	// for _, v := range cells {
	// 	row.AddCell().SetValue(v)
	// }
}

// SetColume 设置列头
// columeName: 列头名，有多少写多少个
func (fd *FileData) SetColume(columeName ...string) {
	if fd.writeSheet == nil {
		fd.writeSheet, _ = fd.AddSheet("newsheet" + strconv.Itoa(len(fd.writeFile.Sheets)+1))
	}
	row := fd.writeSheet.AddRow()
	row.SetHeight(20)
	for _, v := range columeName {
		cell := row.AddCell()
		cell.SetStyle(fd.colStyle)
		cell.SetString(v)
	}
}

// Write 将excel数据写入到writer
// w： io.writer
func (fd *FileData) Write(w io.Writer) error {
	return fd.writeFile.Write(w)
}

// Save 保存excel数据到文件
// 返回保存的完整文件名，错误
func (fd *FileData) Save() (string, error) {
	fn := fd.fileName
	if !strings.HasSuffix(fn, ".xlsx") {
		fn += ".xlsx"
	}
	if !isExist(filepath.Dir(fn)) {
		os.MkdirAll(filepath.Dir(fn), 0o775)
	}

	err := fd.writeFile.Save(fn)
	if err != nil {
		return "", errors.New("excel-文件保存失败:" + err.Error())
	}
	return fn, nil
}

func (fd *FileData) ToFile(f string) (string, error) {
	fn := fd.fileName
	if f != "" {
		fn = f
	}
	if !strings.HasSuffix(fn, ".xlsx") {
		fn += ".xlsx"
	}
	if !isExist(filepath.Dir(fn)) {
		os.MkdirAll(filepath.Dir(fn), 0o775)
	}

	err := fd.writeFile.Save(fn)
	if err != nil {
		return "", errors.New("excel-文件保存失败:" + err.Error())
	}
	return fn, nil
}

// NewExcel 创建新的excel文件
// filename: 需要保存的文件路径，可不加扩展名
// 返回：excel数据格式，错误
func NewExcel(filename string) (*FileData, error) {
	var err error
	fd := &FileData{
		writeFile: xlsx.NewFile(),
		colStyle:  xlsx.NewStyle(),
	}
	fd.colStyle.Alignment.Horizontal = "center"
	fd.colStyle.Font.Bold = true
	fd.colStyle.ApplyAlignment = true
	fd.colStyle.ApplyFont = true
	// fd.writeSheet, err = fd.writeFile.AddSheet(time.Now().Format("2006-01-02"))
	// if err != nil {
	// 	return nil, errors.New("excel-sheet创建失败:" + err.Error())
	// }
	fd.fileName = filename
	return fd, err
}

// NewExcelFromBinary 从文件读取xlsx
func NewExcelFromBinary(bs []byte, filename string) (*FileData, error) {
	var err error
	var fd *FileData
	xf, err := xlsx.OpenBinary(bs)
	if err != nil {
		fd = &FileData{
			writeFile: xlsx.NewFile(),
			colStyle:  xlsx.NewStyle(),
		}
	} else {
		fd = &FileData{
			writeFile: xf,
			colStyle:  xlsx.NewStyle(),
		}
	}
	fd.colStyle.Alignment.Horizontal = "center"
	fd.colStyle.Font.Bold = true
	fd.colStyle.ApplyAlignment = true
	fd.colStyle.ApplyFont = true
	// fd.writeSheet, err = fd.writeFile.AddSheet(time.Now().Format("2006-01-02"))
	// if err != nil {
	// 	return nil, errors.New("excel-sheet创建失败:" + err.Error())
	// }
	fd.fileName = filename
	return fd, err
}

// Merge 合并单元格（适配Row方法返回错误的情况，增加值一致性校验）
// 参数：startRow（起始行）、startCol（起始列）、endRow（结束行）、endCol（结束列）
// 行和列均从0开始
func (fd *FileData) Merge(startRow, startCol, endRow, endCol int) error {
	// 校验sheet是否存在
	if fd.writeSheet == nil {
		return fmt.Errorf("未创建sheet，请先AddRow或SetColume")
	}

	// 校验范围合法性
	if startRow > endRow || startCol > endCol {
		return fmt.Errorf("合并范围无效：起始不能大于结束")
	}
	if startRow == endRow && startCol == endCol {
		return nil // 单个单元格无需合并
	}

	// 1. 获取左上角单元格（基准值）
	topLeftRow, err := fd.writeSheet.Row(startRow)
	if err != nil {
		return fmt.Errorf("获取起始行%d失败：%w", startRow, err)
	}
	topLeftCell := topLeftRow.GetCell(startCol) // 假设Cell方法安全返回单元格（不存在则创建）
	topLeftValue := topLeftCell.Value

	// 2. 值一致性校验（遍历合并范围）
	for r := startRow; r <= endRow; r++ {
		// 获取当前行，处理可能的错误
		row, err := fd.writeSheet.Row(r)
		if err != nil {
			return fmt.Errorf("获取行%d失败：%w", r, err)
		}

		for c := startCol; c <= endCol; c++ {
			// 跳过左上角单元格
			if r == startRow && c == startCol {
				continue
			}

			// 获取当前单元格（假设Cell方法自动创建不存在的单元格）
			cell := row.GetCell(c)
			// 检查值是否一致（非空且不等于基准值则报错）
			if cell.Value != "" && cell.Value != topLeftValue {
				return fmt.Errorf("合并失败：单元格(%d,%d)值为「%s」，与左上角(%d,%d)的「%s」不一致",
					r, c, cell.Value, startRow, startCol, topLeftValue)
			}
		}
	}

	// 3. 执行合并：设置左上角单元格的合并属性
	topLeftCell.HMerge = endCol - startCol
	topLeftCell.VMerge = endRow - startRow

	// 4. 清空合并范围内其他单元格的值
	for r := startRow; r <= endRow; r++ {
		row, err := fd.writeSheet.Row(r)
		if err != nil {
			return fmt.Errorf("获取行%d失败：%w", r, err)
		}

		for c := startCol; c <= endCol; c++ {
			if r == startRow && c == startCol {
				continue // 跳过左上角
			}

			cell := row.GetCell(c)
			cell.SetValue("") // 清空值
			// 保持样式一致（可选）
			if fd.colStyle != nil {
				cell.SetStyle(fd.colStyle)
			}
		}
	}

	return nil
}
func NewExcelFromUpload(file multipart.File, filename string) (*FileData, error) {
	fb, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return NewExcelFromBinary(fb, filename)
}

func isExist(p string) bool {
	if p == "" {
		return false
	}
	_, err := os.Stat(p)
	return err == nil || os.IsExist(err)
}
