package service

import (
	"encoding/csv"
	"io"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/xuri/excelize/v2"
)

const maxUserImportRows = 1000

var userImportHeader = []string{"姓名", "邮箱", "手机号（可选）", "角色（可选）", "初始密码"}

type UserImportRow struct {
	Row      int
	Name     string
	Email    string
	Phone    string
	Role     string
	Password string
}

func ParseUserImportFile(filename string, reader io.Reader) ([]UserImportRow, error) {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".csv":
		return parseUserImportCSV(reader)
	case ".xlsx":
		return parseUserImportXLSX(reader)
	default:
		return nil, errorsx.BadRequest("仅支持 CSV 或 XLSX 文件")
	}
}

func parseUserImportCSV(reader io.Reader) ([]UserImportRow, error) {
	contents, err := io.ReadAll(reader)
	if err != nil {
		return nil, errorsx.BadRequest("无法读取导入文件")
	}
	if !utf8.Valid(contents) {
		return nil, errorsx.BadRequest("CSV 文件必须使用 UTF-8 编码")
	}
	parser := csv.NewReader(strings.NewReader(string(contents)))
	parser.FieldsPerRecord = -1
	records, err := parser.ReadAll()
	if err != nil {
		return nil, errorsx.BadRequest("CSV 文件格式不正确")
	}
	return importRowsFromRecords(records)
}

func parseUserImportXLSX(reader io.Reader) ([]UserImportRow, error) {
	workbook, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, errorsx.BadRequest("XLSX 文件格式不正确")
	}
	defer func() { _ = workbook.Close() }()
	sheets := workbook.GetSheetList()
	if len(sheets) == 0 {
		return nil, errorsx.BadRequest("导入文件缺少工作表")
	}
	iterator, err := workbook.Rows(sheets[0])
	if err != nil {
		return nil, errorsx.BadRequest("无法读取 XLSX 工作表")
	}
	defer func() { _ = iterator.Close() }()
	records := make([][]string, 0)
	for iterator.Next() {
		columns, err := iterator.Columns()
		if err != nil {
			return nil, errorsx.BadRequest("无法读取 XLSX 工作表")
		}
		records = append(records, columns)
	}
	if err := iterator.Error(); err != nil {
		return nil, errorsx.BadRequest("无法读取 XLSX 工作表")
	}
	return importRowsFromRecords(records)
}

func importRowsFromRecords(records [][]string) ([]UserImportRow, error) {
	if len(records) == 0 {
		return nil, errorsx.BadRequest("导入文件缺少表头")
	}
	header := append([]string(nil), records[0]...)
	if len(header) > 0 {
		header[0] = strings.TrimPrefix(header[0], "\uFEFF")
	}
	if !sameImportHeader(header) {
		return nil, errorsx.BadRequest("导入文件表头不正确")
	}
	rows := make([]UserImportRow, 0, len(records)-1)
	for index, record := range records[1:] {
		if importRecordEmpty(record) {
			continue
		}
		if len(rows) >= maxUserImportRows {
			return nil, errorsx.BadRequest("单次最多导入 1000 条数据")
		}
		values := make([]string, len(userImportHeader))
		copy(values, record)
		rows = append(rows, UserImportRow{
			Row: index + 2, Name: values[0], Email: values[1],
			Phone: values[2], Role: values[3], Password: values[4],
		})
	}
	return rows, nil
}

func sameImportHeader(header []string) bool {
	if len(header) != len(userImportHeader) {
		return false
	}
	for index := range userImportHeader {
		if strings.TrimSpace(header[index]) != userImportHeader[index] {
			return false
		}
	}
	return true
}

func importRecordEmpty(record []string) bool {
	for _, value := range record {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}
