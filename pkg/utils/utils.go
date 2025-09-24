package utils

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// ConfirmAction 请求用户确认操作
func ConfirmAction(message string) bool {
	fmt.Printf("%s (y/n) ", message)
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y"
}

// CheckFileAge 检查文件修改时间
func CheckFileAge(filePath string, warnMinutes int) (bool, string) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false, fmt.Sprintf("Failed to stat file: %v", err)
	}

	fileAge := time.Since(fileInfo.ModTime()).Minutes()
	if fileAge > float64(warnMinutes) {
		return true, fmt.Sprintf("File '%s' has not been modified for %.0f minutes", filePath, fileAge)
	}

	return false, ""
}

// FormatTable 格式化表格输出
func FormatTable(headers []string, rows [][]string) string {
	// 计算每列的最大宽度
	columnWidths := make([]int, len(headers))
	for i, header := range headers {
		columnWidths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(columnWidths) && len(cell) > columnWidths[i] {
				columnWidths[i] = len(cell)
			}
		}
	}

	// 构建格式字符串
	formatStr := "  "
	for i, width := range columnWidths {
		formatStr += fmt.Sprintf("%%-%ds", width+2)
		if i < len(columnWidths)-1 {
			formatStr += " "
		}
	}
	formatStr += "\n"

	// 构建表格
	var result strings.Builder

	// 添加标题
	headerRow := make([]interface{}, len(headers))
	for i, h := range headers {
		headerRow[i] = h
	}
	result.WriteString(fmt.Sprintf(formatStr, headerRow...))

	// 添加分隔线
	separatorRow := make([]interface{}, len(headers))
	for i, width := range columnWidths {
		separatorRow[i] = strings.Repeat("-", width)
	}
	result.WriteString(fmt.Sprintf(formatStr, separatorRow...))

	// 添加数据行
	for _, row := range rows {
		rowData := make([]interface{}, len(row))
		for i, cell := range row {
			rowData[i] = cell
		}
		result.WriteString(fmt.Sprintf(formatStr, rowData...))
	}

	return result.String()
}

// MaskSensitiveInfo 隐藏敏感信息
func MaskSensitiveInfo(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + strings.Repeat("*", len(s)-4) + s[len(s)-2:]
}

// FormatFileSize 格式化文件大小
func FormatFileSize(size int64) string {
	const (
		B  = 1
		KB = 1024 * B
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}
