package exportcsv

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/xuri/excelize/v2"
)

var numericCellPattern = regexp.MustCompile(`^[+-]?\d+(?:\.\d+)?$`)

func buildXLSX(csvData []byte, scope Scope, period Period, lang Language, partial bool) ([]byte, error) {
	trimmed := bytes.TrimPrefix(csvData, []byte("\ufeff"))
	reader := csv.NewReader(bytes.NewReader(trimmed))
	reader.Comma = delimiterForLanguage(lang)
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, errors.New("no records available for export")
	}

	headers := records[0]
	dataRows := records[1:]

	file := excelize.NewFile()
	sheetName := scopeSheetName(scope, lang)
	defaultSheet := file.GetSheetName(0)
	if defaultSheet != sheetName {
		if err := file.SetSheetName(defaultSheet, sheetName); err != nil {
			return nil, err
		}
	}

	titleStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#FFFFFF", Size: 14},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#1D4ED8"}, Pattern: 1},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
	})
	if err != nil {
		return nil, err
	}

	subtitleStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "#64748B", Size: 10},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
	})
	if err != nil {
		return nil, err
	}

	headerStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#0F172A"}, Pattern: 1},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#334155", Style: 1},
			{Type: "right", Color: "#334155", Style: 1},
			{Type: "top", Color: "#334155", Style: 1},
			{Type: "bottom", Color: "#334155", Style: 1},
		},
	})
	if err != nil {
		return nil, err
	}

	bodyStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "#0F172A"},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#E2E8F0", Style: 1},
			{Type: "right", Color: "#E2E8F0", Style: 1},
			{Type: "top", Color: "#E2E8F0", Style: 1},
			{Type: "bottom", Color: "#E2E8F0", Style: 1},
		},
	})
	if err != nil {
		return nil, err
	}

	numberStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "#0F172A"},
		Alignment: &excelize.Alignment{
			Horizontal: "right",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#E2E8F0", Style: 1},
			{Type: "right", Color: "#E2E8F0", Style: 1},
			{Type: "top", Color: "#E2E8F0", Style: 1},
			{Type: "bottom", Color: "#E2E8F0", Style: 1},
		},
		CustomNumFmt: ptr("0.##"),
	})
	if err != nil {
		return nil, err
	}

	title := fmt.Sprintf("%s - %s", text(lang, "Export Data", "Export Data"), scopeLabel(scope, lang))
	subtitle := exportPeriodSummary(period, lang)
	if partial {
		subtitle = subtitle + " • " + text(lang, "Sebagian data tidak tersedia", "Some data was skipped")
	}

	lastCol, err := excelize.ColumnNumberToName(len(headers))
	if err != nil {
		return nil, err
	}

	if err := file.SetCellValue(sheetName, "A1", title); err != nil {
		return nil, err
	}
	if err := file.MergeCell(sheetName, "A1", fmt.Sprintf("%s1", lastCol)); err != nil {
		return nil, err
	}
	if err := file.SetCellStyle(sheetName, "A1", fmt.Sprintf("%s1", lastCol), titleStyle); err != nil {
		return nil, err
	}

	if err := file.SetCellValue(sheetName, "A2", subtitle); err != nil {
		return nil, err
	}
	if err := file.MergeCell(sheetName, "A2", fmt.Sprintf("%s2", lastCol)); err != nil {
		return nil, err
	}
	if err := file.SetCellStyle(sheetName, "A2", fmt.Sprintf("%s2", lastCol), subtitleStyle); err != nil {
		return nil, err
	}

	headerRow := 4
	dataStartRow := 5
	columnWidths := make([]float64, len(headers))

	for colIndex, header := range headers {
		cell, err := excelize.CoordinatesToCellName(colIndex+1, headerRow)
		if err != nil {
			return nil, err
		}
		if err := file.SetCellValue(sheetName, cell, header); err != nil {
			return nil, err
		}
		if err := file.SetCellStyle(sheetName, cell, cell, headerStyle); err != nil {
			return nil, err
		}
		columnWidths[colIndex] = measuredWidth(header)
	}

	for rowIndex, record := range dataRows {
		excelRow := dataStartRow + rowIndex
		for colIndex := range headers {
			value := ""
			if colIndex < len(record) {
				value = record[colIndex]
			}

			cell, err := excelize.CoordinatesToCellName(colIndex+1, excelRow)
			if err != nil {
				return nil, err
			}

			if isNumericCell(value) {
				numberValue, _ := strconv.ParseFloat(value, 64)
				if err := file.SetCellValue(sheetName, cell, numberValue); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, numberStyle); err != nil {
					return nil, err
				}
			} else {
				if err := file.SetCellValue(sheetName, cell, value); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, bodyStyle); err != nil {
					return nil, err
				}
			}

			if width := measuredWidth(value); width > columnWidths[colIndex] {
				columnWidths[colIndex] = width
			}
		}
	}

	for colIndex, width := range columnWidths {
		columnName, err := excelize.ColumnNumberToName(colIndex + 1)
		if err != nil {
			return nil, err
		}
		if width < 12 {
			width = 12
		}
		if width > 42 {
			width = 42
		}
		if err := file.SetColWidth(sheetName, columnName, columnName, width+2); err != nil {
			return nil, err
		}
	}

	if err := file.SetRowHeight(sheetName, 1, 24); err != nil {
		return nil, err
	}
	if err := file.SetRowHeight(sheetName, 2, 20); err != nil {
		return nil, err
	}
	if err := file.SetRowHeight(sheetName, headerRow, 22); err != nil {
		return nil, err
	}
	if err := file.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      headerRow,
		TopLeftCell: fmt.Sprintf("A%d", dataStartRow),
		ActivePane:  "bottomLeft",
		Selection: []excelize.Selection{
			{SQRef: fmt.Sprintf("A%d", dataStartRow), Pane: "bottomLeft"},
		},
	}); err != nil {
		return nil, err
	}

	sheetIndex, err := file.GetSheetIndex(sheetName)
	if err != nil {
		return nil, err
	}
	file.SetActiveSheet(sheetIndex)

	if err := file.AutoFilter(sheetName, fmt.Sprintf("A%d:%s%d", headerRow, lastCol, headerRow), []excelize.AutoFilterOptions{}); err != nil {
		return nil, err
	}

	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func scopeLabel(scope Scope, lang Language) string {
	switch scope {
	case ScopeTransactions:
		return text(lang, "Transaksi", "Transactions")
	case ScopeDebts:
		return text(lang, "Utang", "Debts")
	case ScopeReports:
		return text(lang, "Laporan", "Reports")
	default:
		return string(scope)
	}
}

func scopeSheetName(scope Scope, lang Language) string {
	switch scope {
	case ScopeTransactions:
		return text(lang, "Transaksi", "Transactions")
	case ScopeDebts:
		return text(lang, "Utang", "Debts")
	case ScopeReports:
		return text(lang, "Laporan", "Reports")
	default:
		return "Export"
	}
}

func exportPeriodSummary(period Period, lang Language) string {
	if period.Label != "" {
		label := strings.ReplaceAll(period.Label, "_to_", " - ")
		return fmt.Sprintf("%s: %s", text(lang, "Periode", "Period"), label)
	}

	return text(lang, "Semua data", "All data")
}

func measuredWidth(value string) float64 {
	if value == "" {
		return 0
	}

	// Use rune count to better estimate width for localized text.
	return float64(utf8.RuneCountInString(value))
}

func isNumericCell(value string) bool {
	return numericCellPattern.MatchString(strings.TrimSpace(value))
}

func ptr(value string) *string {
	return &value
}
