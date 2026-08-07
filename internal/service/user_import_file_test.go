package service

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestParseUserImportCSVPreservesSourceRows(t *testing.T) {
	input := "姓名,邮箱,手机号（可选）,角色（可选）,初始密码\n" +
		"张三,ZHANG@example.com,,学员,password1\n" +
		",,,,\n" +
		"李四,li@example.com,13800138000,instructor,password2\n"

	rows, err := ParseUserImportFile("users.CSV", strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseUserImportFile() error = %v", err)
	}
	if len(rows) != 2 || rows[0].Row != 2 || rows[1].Row != 4 {
		t.Fatalf("rows = %#v, want source rows 2 and 4", rows)
	}
	if rows[0].Name != "张三" || rows[0].Email != "ZHANG@example.com" ||
		rows[1].Phone != "13800138000" || rows[1].Role != "instructor" {
		t.Fatalf("rows = %#v, fields were not preserved", rows)
	}
}

func TestParseUserImportXLSXPreservesSourceRows(t *testing.T) {
	file := excelize.NewFile()
	t.Cleanup(func() { _ = file.Close() })
	sheet := file.GetSheetName(0)
	rows := [][]interface{}{
		{"姓名", "邮箱", "手机号（可选）", "角色（可选）", "初始密码"},
		{"张三", "zhang@example.com", "", "学员", "password1"},
		{"", "", "", "", ""},
		{"李四", "li@example.com", "13800138000", "instructor", "password2"},
	}
	for index, row := range rows {
		cell, err := excelize.CoordinatesToCellName(1, index+1)
		if err != nil {
			t.Fatal(err)
		}
		if err := file.SetSheetRow(sheet, cell, &row); err != nil {
			t.Fatal(err)
		}
	}
	var contents bytes.Buffer
	if err := file.Write(&contents); err != nil {
		t.Fatal(err)
	}

	parsed, err := ParseUserImportFile("users.xlsx", bytes.NewReader(contents.Bytes()))
	if err != nil {
		t.Fatalf("ParseUserImportFile() error = %v", err)
	}
	if len(parsed) != 2 || parsed[0].Row != 2 || parsed[1].Row != 4 {
		t.Fatalf("rows = %#v, want source rows 2 and 4", parsed)
	}
}

func TestParseUserImportXLSXRejectsInvalidHeader(t *testing.T) {
	file := excelize.NewFile()
	t.Cleanup(func() { _ = file.Close() })
	sheet := file.GetSheetName(0)
	row := []interface{}{"姓名", "邮箱"}
	if err := file.SetSheetRow(sheet, "A1", &row); err != nil {
		t.Fatal(err)
	}
	var contents bytes.Buffer
	if err := file.Write(&contents); err != nil {
		t.Fatal(err)
	}

	if _, err := ParseUserImportFile("users.xlsx", bytes.NewReader(contents.Bytes())); errorCode(err) != 40000 {
		t.Fatalf("ParseUserImportFile() error = %#v, want bad request", err)
	}
}

func TestParseUserImportCSVRejectsInvalidFiles(t *testing.T) {
	header := "姓名,邮箱,手机号（可选）,角色（可选）,初始密码\n"
	tests := []struct {
		name     string
		filename string
		input    string
	}{
		{name: "extension", filename: "users.txt", input: header},
		{name: "header", filename: "users.csv", input: "姓名,邮箱\n"},
		{name: "encoding", filename: "users.csv", input: header + string([]byte{0xff, 0xfe})},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ParseUserImportFile(test.filename, strings.NewReader(test.input)); errorCode(err) != 40000 {
				t.Fatalf("ParseUserImportFile() error = %#v, want bad request", err)
			}
		})
	}
}

func TestParseUserImportCSVRejectsMoreThanOneThousandRows(t *testing.T) {
	var input strings.Builder
	input.WriteString("姓名,邮箱,手机号（可选）,角色（可选）,初始密码\n")
	for index := 0; index < 1001; index++ {
		fmt.Fprintf(&input, "用户%d,user%d@example.com,,,password1\n", index, index)
	}

	if _, err := ParseUserImportFile("users.csv", strings.NewReader(input.String())); errorCode(err) != 40000 {
		t.Fatalf("ParseUserImportFile() error = %#v, want bad request", err)
	}
}
