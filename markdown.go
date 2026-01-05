package cml

import "strings"

// ToMarkdown 将基元序列还原为人类可读的 Markdown 格式
func ToMarkdown(elements []*CmlElement) string {
	var sb strings.Builder
	for _, el := range elements {
		if el.Type == TypeToken {
			// 处理内嵌反引号的情况
			wrap := "`"
			if strings.Contains(el.Value, "`") {
				wrap = "```"
			}
			sb.WriteString(wrap + el.Value + wrap)
		} else {
			sb.WriteString(el.Value)
		}
	}
	return sb.String()
}
