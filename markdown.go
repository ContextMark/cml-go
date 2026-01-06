package cml

/**
Markdown转换不列入语法内核:
1、视为上层内置的方法，UI编辑器和语法编码器应该分层
2、也不适用大规模计算，也没必要
3、同时也可以提供一个反向的方法，剔除换行、多空格等规范写法
*/
import (
	"cml/internal"
	"strings"
)

// 将原始的 CML 编码字符串还原为人类可读的 Markdown（行内代码风格）
func ToMarkdown(encoded string) string {
	var sb strings.Builder

	//先反解成双序列
	fragments, err := internal.CML2Fragments(encoded)
	if err != nil {
		return ""
	}

	for i := 0; i < len(fragments.Tokens); i++ {
		// 1. 写入 Token（行内代码，符合 CommonMark）
		sb.WriteString(wrapInlineCode(fragments.Tokens[i]))

		// 2. 写入关系符（最后一个 Token 后没有关系符）
		if i < len(fragments.Relations) {
			sb.WriteString(fragments.Relations[i])
		}
	}

	return sb.String()
}

// wrapInlineCode 使用“比内容中连续反引号更多的反引号”包裹行内代码
// CommonMark 规则：
// - 行内代码可以用任意数量的反引号包裹
// - 包裹符数量必须 > 内容中最长连续反引号数量
// 输入ab，输出`ab`，输入a`b，输出```a`b```
func wrapInlineCode(s string) string {
	maxRun := 0 // 记录全局发现过的“最长连续长度”
	cur := 0    // 记录当前正在数着的“连续长度”

	for _, r := range s {
		//如果有反引号字符字符
		if r == '`' {
			cur++
			if cur > maxRun {
				maxRun = cur
			}
		} else {
			cur = 0
		}
	}

	wrap := strings.Repeat("`", maxRun+1)
	// 如果要包裹的内容以 ` 开头或结尾，或者为了安全起见统一加空格
	// 规范要求：如果加了空格，首尾必须成对加，渲染器会自动剔除首尾各一个空格
	if strings.HasPrefix(s, "`") || strings.HasSuffix(s, "`") {
		return wrap + " " + s + " " + wrap
	}
	return wrap + s + wrap
}

// FromMarkdown 将人类可读的 Markdown 还原为 CML 编码前的原始字符串
// 处理规则：
// 1. 所有空白字符（连续空格 / 换行 / tab）统一压缩为单个空格
// 2. 解析并移除行内 code 包裹（支持任意数量反引号）
// 3. 不支持代码块（``` 独占一行的情况）——视为非法或普通文本
func FromMarkdown(md string) string {
	// 1. 统一空白字符为单空格
	normalized := normalizeWhitespace(md)

	// 2. 解析行内 code，剥离反引号
	var sb strings.Builder
	n := len(normalized)

	for i := 0; i < n; {
		//如果不是反引号，直接写
		if normalized[i] != '`' {
			sb.WriteByte(normalized[i])
			i++
			continue
		}

		// 如果是反引号，还要统计包裹反引号数量
		j := i
		for j < n && normalized[j] == '`' { //下一个不是反引号就中断检查
			j++
		}
		wrapLen := j - i

		// 进一步查找对应结束包裹
		k := j
		for k < n {
			//如果是反引号，就开始统计
			if normalized[k] == '`' {
				t := k
				for t < n && normalized[t] == '`' {
					t++
				}
				if t-k == wrapLen {
					// 只有数量相等才命中结束包裹
					sb.WriteString(normalized[j:k])
					i = t
					goto next //直接跳出去
				}
				k = t //没有命中就重新计数
			} else {
				k++
			}
		}

		// 找不到匹配的结束包裹，就视为普通字符
		sb.WriteByte(normalized[i])
		i++

	next:
	}

	return strings.TrimSpace(sb.String())
}

// 将任意连续空白字符压缩为单个空格
func normalizeWhitespace(s string) string {
	var sb strings.Builder
	space := false

	for _, r := range s {
		switch r {
		case ' ', '\n', '\r', '\t': //空格、换行、制表符
			//只有第一个空白才使用空格，其余直接放过去
			if !space {
				sb.WriteByte(' ')
				space = true
			}
		default:
			sb.WriteRune(r)
			space = false
		}
	}

	return sb.String()
}
