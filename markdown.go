package cml

/**
Markdown转换不列入语法内核:
1、视为上层内置的方法，UI编辑器和语法编码器应该分层
2、也不适用大规模计算，也没必要
3、同时也可以提供一个反向的方法，剔除换行、多空格等规范写法
*/
import (
	"errors"
	"fmt"
	"strings"

	"github.com/ContextMark/cml-go/internal"
)

// 将原始的 CML 编码字符串还原为人类可读的 Markdown（行内代码风格）
func toMarkdown(encoded string) string {
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

// fromMarkdown 将人类可读的 Markdown 还原为 CML 编码前的原始字符串
// 严格按照：1.全局压缩 -> 2.识别反引号包裹的Token -> 3.识别单字符关系符
// 处理规则：
// 1. 先一刀切，将连续空格 / 换行 / tab，需要反向统一压缩为单个空格，在语义上通常没有意义
// 2、按反引号包裹token+连接单字符关系符的规则进行切分，解析出基元序列
// 3. 确保不同token之间的关系分隔符，都是单字符
// 4. 识别用任意数量反引号包括token内部反引号的情况，因为冲突，所以不需要支持代码块（``` 独占一行的情况）——视为非法或普通文本，跨行的代码块可以视为一个token
// FromMarkdown 严格执行：全局空白压缩 -> 识别反引号序列 -> 提取单字符关系符
func fromMarkdown(md string) ([]string, error) {
	// --- 第一步：执行“一刀切”预处理 ---
	// 根据规则：将连续的空格、换行、Tab 统一压缩为单个空格，消除无意义的格式噪声
	normalized := normalizeWhitespace(md)

	var result []string
	var current strings.Builder
	const seps = "@.+: " // CML 规范定义的 5 类合法单字符关系符
	n := len(normalized)

	// commitToken 是一个内部辅助闭包，用于将当前积累的普通文本提交为 Token
	commitToken := func() {
		if current.Len() > 0 {
			result = append(result, current.String())
			current.Reset()
		}
	}

	// --- 第二步：流式扫描解析 ---
	for i := 0; i < n; {
		// 场景 A：处理反引号包裹 (优先级最高，用于保护内部包含分隔符的 Token)
		if normalized[i] == '`' {
			// 1. 计算起始反引号的数量 (支持 `` 或 ``` 等任意长度)
			j := i
			for j < n && normalized[j] == '`' {
				j++
			}
			wrapLen := j - i

			// 2. 寻找匹配的闭合反引号组
			foundMatch := false
			for k := j; k <= n-wrapLen; k++ {
				// 匹配条件：长度一致且其后不再紧跟反引号
				if normalized[k:k+wrapLen] == normalized[i:j] {
					if k+wrapLen == n || normalized[k+wrapLen] != '`' {
						commitToken()                            // 提交包裹前的普通文本
						result = append(result, normalized[j:k]) // 将包裹内的原始内容作为独立 Token
						i = k + wrapLen
						foundMatch = true
						break
					}
				}
			}

			// 【异常处理】：如果反引号未闭合，视为语法错误
			if !foundMatch {
				return nil, fmt.Errorf("语法错误: 索引 %d 处发现未闭合的反引号", i)
			}
			continue
		}

		// 场景 B：处理单字符关系分隔符
		char := normalized[i]
		if strings.ContainsRune(seps, rune(char)) {
			commitToken()                         // 提交分隔符左侧的文本
			result = append(result, string(char)) // 关系符本身作为独立元素
			i++
			continue
		}

		// 场景 C：普通字符积累 (形成 Token 的一部分)
		current.WriteByte(char)
		i++
	}
	commitToken() // 提交末尾剩余的文本

	// --- 第三步：CML 物理结构严格校验 ---
	// 规则：CML 必须是 <Token> <Relation> <Token> 的交替结构

	count := len(result)
	if count == 0 {
		return nil, errors.New("解析失败: 输入内容未包含有效基元")
	}

	// 1. 长度校验：交替序列的元素总数必须是奇数
	if count%2 == 0 {
		return nil, errors.New("语法错误: CML 序列不完整，不能以关系符开头或结尾")
	}

	// 2. 序列内容校验：遍历检查每一位的物理属性
	for i, val := range result {
		if i%2 != 0 {
			// 奇数位 (1, 3, 5...) 必须是单字符关系符
			if len(val) != 1 || !strings.ContainsAny(val, seps) {
				return nil, fmt.Errorf("语法错误: 位置 %d 期望关系符，实际得到 '%s'", i, val)
			}
		} else {
			// 偶数位 (0, 2, 4...) 必须是 Token
			// 额外检查：确保 Token 位（即当前关系符的前后）不是空的
			if val == "" {
				return nil, fmt.Errorf("语法错误:Token不应该是空串 (索引 %d)", i)
			}
		}
	}
	return result, nil
}

// normalizeWhitespace 将所有连续的空白字符（空格、换行、制表符）压缩为单个空格
func normalizeWhitespace(s string) string {
	var sb strings.Builder
	inSpace := false // 标记当前是否正处于空白序列中

	for _, r := range s {
		switch r {
		case ' ', '\n', '\r', '\t':
			// 如果是空白序列的第一个字符，则写入空格；后续连续的直接跳过
			if !inSpace {
				sb.WriteByte(' ')
				inSpace = true
			}
		default:
			sb.WriteRune(r)
			inSpace = false // 遇到非空白字符，重置标记
		}
	}
	return strings.TrimSpace(sb.String()) // 移除整体首尾可能残留的空格
}
