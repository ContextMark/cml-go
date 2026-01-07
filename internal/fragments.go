package internal

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/shengdoushi/base58"
)

// New 将平铺的字符串切片转换为双列表结构的 CmlFragments
// 输入示例: []string{"user", "@", "domain", ".", "com"}
// 输出示例: Tokens: ["user", "domain", "com"], Relations: ["@", "."]
func New(arr []string) (*CmlFragments, error) {
	size := len(arr)

	// 1. 基础校验：CML 物理特性要求序列必须是奇数 (T R T R T)
	if size == 0 {
		return nil, errors.New("切片不能为空")
	}
	if size%2 == 0 {
		return nil, errors.New("CML 序列长度必须为奇数（Token 与 关系符 必须交替出现）")
	}

	// 2. 预分配内存：已知长度，直接分配以提升性能
	tokenCount := (size / 2) + 1
	relationCount := size / 2

	fragments := &CmlFragments{
		Tokens:    make([]string, 0, tokenCount),
		Relations: make([]string, 0, relationCount),
	}

	// 3. 拆解序列
	for i, val := range arr {
		if i%2 == 0 {
			// 偶数索引一定是 Token
			fragments.Tokens = append(fragments.Tokens, val)
		} else {
			// 奇数索引一定是 Relation
			const seps = "@.+: "
			// CML 规范定义的五个合法符：@, ., +, :, 空格
			if len(val) != 1 || !strings.ContainsAny(val, seps) {
				return nil, fmt.Errorf("语法错误: CML禁止以关系符开头:'%s'", val)
			}
			fragments.Relations = append(fragments.Relations, val)
		}
	}

	return fragments, nil
}

/**
--- 一、基元序列检查 ---
*/

// 检查 CmlFragments 是否符合奇偶性特征
// CML 规则：Token 数量必须比 Relations 数量多 1 (即：T R T R T)
func (f *CmlFragments) IsValid() error {
	if f == nil {
		return errors.New("CML基元双序列不应为nil")
	}
	tLen := len(f.Tokens)
	rLen := len(f.Relations)

	if tLen == 0 {
		return errors.New("CML基元双序列的token序列不应为空")
	}
	if tLen != rLen+1 {
		return fmt.Errorf("CML基元双序列组成不合法: Token数量(%d)与关系符数量(%d)不匹配", tLen, rLen)
	}
	return nil
}

/**
------------------------ 编码类方法 --------------------------
*/

// EncodeA 编码成 a 模式：双层 base58
// 字符集普适性最好，但性能不理想，适用于对字符集安全性要求极高的场景
func (f *CmlFragments) EncodeA() (string, error) {
	if err := f.IsValid(); err != nil {
		return "", err
	}

	var sb strings.Builder
	// 遍历 Tokens 和 Relations 进行交替拼接
	for i := 0; i < len(f.Tokens); i++ {
		// Token 级 base58 编码
		sb.WriteString(base58.Encode([]byte(f.Tokens[i]), base58.BitcoinAlphabet))

		// 拼接触点关系符（最后一个 Token 后没有关系符）
		if i < len(f.Relations) {
			sb.WriteString(f.Relations[i])
		}
	}

	// 整体进行二次 base58 编码
	payload := base58.Encode([]byte(sb.String()), base58.BitcoinAlphabet)
	return "a" + payload, nil
}

// EncodeC 编码成 c 模式：双层 base64 (高性能)
// 适合大规模、高并发处理场景
func (f *CmlFragments) EncodeC() (string, error) {
	if err := f.IsValid(); err != nil {
		return "", err
	}

	var sb strings.Builder
	for i := 0; i < len(f.Tokens); i++ {
		// Token 级 base64 编码
		sb.WriteString(base64.RawURLEncoding.EncodeToString([]byte(f.Tokens[i])))
		// 拼接触点关系符（最后一个 Token 后没有关系符）
		if i < len(f.Relations) {
			sb.WriteString(f.Relations[i])
		}
	}

	// 整体进行二次 base64 编码
	payload := base64.RawURLEncoding.EncodeToString([]byte(sb.String()))
	return "c" + payload, nil
}

// EncodeP 编码成 p 模式：单层明文混编
// 保持最小熵增，提供最佳的可读性与长度比
func (f *CmlFragments) EncodeP() (string, error) {
	if err := f.IsValid(); err != nil {
		return "", err
	}
	return "p" + f.buildBase64URLPayload(), nil
}

// EncodeQ 编码成 q 模式：双层混编
// 在不可读的前提下，通过智能判断减少不必要的 Base64 转换，提供最小熵增
func (f *CmlFragments) EncodeQ() (string, error) {
	if err := f.IsValid(); err != nil {
		return "", err
	}
	payload := f.buildBase64URLPayload()
	return "q" + base64.RawURLEncoding.EncodeToString([]byte(payload)), nil
}

/**
--- 辅助逻辑：内部 Payload 构建 ---
*/

// 处理 P 和 Q 模式共用的混编逻辑
// 如果 Token 包含保留字符，则进行 Base64 处理并标记索引
func (f *CmlFragments) buildBase64URLPayload() string {
	var sb strings.Builder

	//检查每一个token，对包含保留字符集的进行强制编码
	for i := 0; i < len(f.Tokens); i++ {
		token := f.Tokens[i]

		// 调用独立函数处理单个 token
		sb.WriteString(processToken(token))

		// 拼接关系符，关系符数量少1
		if i < len(f.Relations) {
			sb.WriteString(f.Relations[i])
		}
	}
	return sb.String()
}

/*
*
检查单个Token 是否包含关系符，需要编码:
1、range遍历的是 rune（字节级的Unicode码点），而不是字符
2、对于纯 ASCII（0~127）的保留字符，rune 值就是 byte 值，可以直接比较，这样性能最高
需要编码规避的保留字符集（"@.+: !"）：5关系符+编码转义符！
*/
func processToken(token string) string {
	for _, b := range token { // 字节遍历
		//单引号表示rune类型，不能用双引号哦
		if b == '@' || b == '.' || b == '+' || b == ':' || b == ' ' || b == '!' {
			// 对包含敏感字符的 Token 进行编码，并追加 '!' 标记
			return base64.RawURLEncoding.EncodeToString([]byte(token)) + "!"
		}
	}
	// 不需要编码，直接返回原 token
	return token
}
