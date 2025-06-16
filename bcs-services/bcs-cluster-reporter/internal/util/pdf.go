/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package util xxx
package util

import (
	"math"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

// PDFTable xxx
type PDFTable struct {
	Data   [][]Column
	Header []Column
	Title  Column
	Line   int
}

// Column xxx
type Column struct {
	Content string
	Color   Color
}

// Color xxx
type Color struct {
	Red   int
	Green int
	Blue  int
}

const (
	maxRowLineNum   = 5
	minRowLineWidth = 80
)

// WritePDFTable xxx
func WritePDFTable(pdf *gofpdf.Fpdf, pdfTable PDFTable, wrap bool) {
	pdf.SetTextColor(0, 0, 0) // 设置表格文本颜色
	pdf.SetDrawColor(0, 0, 0) // 设置表格边框颜色
	pdf.SetLineWidth(0.2)     // 设置表格边框宽度

	// 打印标题
	WritePDFTableTtile(pdf, pdfTable.Title)

	// 打印tablebody
	WritePDFTableBody(pdf, pdfTable.Header, pdfTable.Data)

	// 与后面的表格隔开
	if wrap {
		pdf.Ln(-1)
	}
}

// WritePDFTableTtile xxx
func WritePDFTableTtile(pdf *gofpdf.Fpdf, title Column) {
	pdf.SetDrawColor(0, 0, 0) // 设置表格边框颜色
	pdf.SetLineWidth(0.2)     // 设置表格边框宽度
	pdf.SetFont("tencent", "", 12)

	pageWidth, _ := pdf.GetPageSize()
	// 打印标题
	y := pdf.GetY()

	pdf.SetXY(0, y)

	pdf.SetTextColor(title.Color.Red, title.Color.Green, title.Color.Blue) // 设置表格文本颜色
	pdf.MultiCell(pageWidth, 10, title.Content, "0", "L", false)
}

// WritePDFTableBody xxx
func WritePDFTableBody(pdf *gofpdf.Fpdf, headers []Column, rowList [][]Column) {
	pdf.SetTextColor(0, 0, 0) // 设置表格文本颜色
	pdf.SetDrawColor(0, 0, 0) // 设置表格边框颜色
	pdf.SetLineWidth(0.2)     // 设置表格边框宽度

	pageWidth, pageHeight := pdf.GetPageSize()
	var lineWidth = pageWidth - 10
	var lineHeight float64 = 5
	columnWidthList := GetcolumnWidthList(append(rowList, headers), lineWidth, pdf)

	// 计算当前页是否足够打印
	var lineNumSum float64
	for _, row := range rowList {
		var lines float64 = 1
		for index, value := range row {
			relLines := getStringLines(value.Content, columnWidthList[index], pdf)
			if relLines > lines {
				lines = relLines
			}
		}
		lineNumSum = lineNumSum + 1
	}

	y := pdf.GetY()
	pageNo := pdf.PageNo()
	if (pageHeight - y) < (lineNumSum * lineHeight) {
		pdf.SetPage(pageNo + 1)
	}

	startX := (pageWidth - lineWidth) / 2
	if startX < 0 {
		startX = 0
	}

	// 设置表头
	pdf.SetFont("tencent", "", 12)
	pdf.SetFillColor(182, 215, 228) // 设置表格背景颜色
	for index, header := range headers {
		// 防止自动换页时，每一列都另起一页
		if pdf.PageNo() > pageNo {
			pageNo = pdf.PageNo()
			pdf.SetPage(pageNo)
			_, top, _, _ := pdf.GetMargins()
			y = top
		}

		pdf.SetXY(startX, y)
		pdf.SetTextColor(header.Color.Red, header.Color.Green, header.Color.Blue) // 设置表格文本颜色
		pdf.MultiCell(columnWidthList[index], lineHeight, header.Content, "1", "L", true)
		startX = startX + columnWidthList[index]

	}

	// 打印表格
	pdf.SetFont("tencent", "", 12)
	pdf.SetFillColor(240, 240, 240) // 设置表格背景颜色
	for _, row := range rowList {
		var lines float64 = 1
		for index, value := range row {
			relLines := getStringLines(value.Content, columnWidthList[index], pdf)
			if relLines > lines {
				lines = relLines
			}
		}

		startX = (pageWidth - lineWidth) / 2
		y = pdf.GetY()
		for index, value := range row {
			// 防止自动换页时，每一列都另起一页
			if pdf.PageNo() > pageNo {
				pageNo = pdf.PageNo()
				pdf.SetPage(pageNo)
				_, top, _, _ := pdf.GetMargins()
				y = top
			}

			if value.Content == "" {
				value.Content = " "
			}
			pdf.SetXY(startX, y)
			pdf.SetTextColor(value.Color.Red, value.Color.Green, value.Color.Blue) // 设置表格文本颜色
			pdf.MultiCell(columnWidthList[index],
				lineHeight*(lines/getStringLines(value.Content, columnWidthList[index], pdf)),
				value.Content, "1", "L", true)
			startX = startX + columnWidthList[index]
		}
	}
}

// WriteHorizontalPDFTableLine xxx
func WriteHorizontalPDFTableLine(columnList []Column, columnWidthList []float64, pdf *gofpdf.Fpdf) {
	y := pdf.GetY()
	pageWidth, _ := pdf.GetPageSize()
	lineWidth := pageWidth - 10
	startX := (pageWidth - lineWidth) / 2
	var lineHeight float64 = 5

	var lines float64 = 1
	for index, value := range columnList {
		relLines := getStringLines(value.Content, columnWidthList[index], pdf)
		if relLines > lines {
			lines = relLines
		}
	}

	pageNo := pdf.PageNo()
	for index, value := range columnList {
		// 防止自动换页时，每一列都另起一页
		if pdf.PageNo() > pageNo {
			pageNo = pdf.PageNo()
			pdf.SetPage(pageNo)
			_, top, _, _ := pdf.GetMargins()
			y = top
		}

		if value.Content == "" {
			value.Content = " "
		}
		pdf.SetXY(startX, y)
		if (index % 2) == 0 {
			pdf.SetFont("tencent", "", 12)
			pdf.SetFillColor(182, 215, 228)
			pdf.SetTextColor(value.Color.Red, value.Color.Green, value.Color.Blue) // 设置表格文本颜色
			pdf.MultiCell(columnWidthList[index], lineHeight*(lines/getStringLines(value.Content, columnWidthList[index], pdf)), value.Content, "1", "L", true)

		} else {
			pdf.SetFont("tencent", "", 12)
			pdf.SetFillColor(240, 240, 240)
			pdf.SetTextColor(value.Color.Red, value.Color.Green, value.Color.Blue) // 设置表格文本颜色
			pdf.MultiCell(columnWidthList[index], lineHeight*(lines/getStringLines(value.Content, columnWidthList[index], pdf)), value.Content, "1", "L", false)
		}
		startX = startX + columnWidthList[index]
	}
}

// 判断该字符串最终会被打印成多少行
func getStringLines(str string, width float64, pdf *gofpdf.Fpdf) float64 {
	var relLines float64
	for _, line := range strings.Split(str, "\n") {
		addLines := math.Ceil((pdf.GetStringWidth(line)) / (width - 2)) // 每行实际上还有margin的宽度
		if addLines == 0 {
			addLines = 1
		}
		relLines = relLines + addLines
	}
	return relLines
}

// GetcolumnWidthList xxx
func GetcolumnWidthList(lines [][]Column, lineWidth float64, pdf *gofpdf.Fpdf) []float64 {
	maxWidth := lineWidth
	var columnWidthList = make([]float64, len(lines[0]), len(lines[0]))
	for _, row := range lines {
		for index, value := range row {
			// 增加5宽度做margin和冗余
			if columnWidthList[index] < pdf.GetStringWidth(value.Content)+2 {
				columnWidthList[index] = pdf.GetStringWidth(value.Content) + 2
			}
			// 至少需要留给后面几列的空间
			if columnWidthList[index] > (maxWidth - float64(len(row)-index-1)*30) {
				columnWidthList[index] = maxWidth - float64(len(row)-index-1)*30
			}
			maxWidth = maxWidth - columnWidthList[index]
		}
		maxWidth = lineWidth
	}

	// 平均分配剩余的width
	var columnWidthSum float64
	for _, columnWidth := range columnWidthList {
		columnWidthSum = columnWidthSum + columnWidth
	}

	if lineWidth > columnWidthSum {
		for index, columnWidth := range columnWidthList {
			columnWidthList[index] = columnWidth + (lineWidth-columnWidthSum)/float64(len(columnWidthList))
		}
	}

	var columnWidthSum1 float64
	for _, columnWidth := range columnWidthList {
		columnWidthSum1 = columnWidthSum1 + columnWidth
	}

	return columnWidthList
}
