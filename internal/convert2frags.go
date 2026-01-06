package internal

/**
--- 二、基元和CML字符串互转 ---
*/
import (
	"fmt"
	"strings"
)

// CML2Elements 将编码后的CML字符串解析为基元序列
func CML2Fragments(encoded string) (*CmlFragments, error) {
	mode, rawPayload, err := decodePayload(encoded)
	if err != nil {
		return nil, err
	}
	return parseRawPayload2Double(rawPayload, mode)
}

// 解析解密后的原始荷载字符串，转换成两类基元的双序列
func parseRawPayload2Double(raw string, mode uint8) (*CmlFragments, error) {
	n := len(raw)
	if n == 0 {
		return nil, fmt.Errorf("非法的空CML")
	}
	// 定义5个分隔符集合
	const seps = "@.+: "
	/**
	关系符必然不出现在首位和尾位
	- 不允许 @@
	- 不允许 A@
	- 不允许 @B
	- 不允许 A..@@..B
	*/
	if strings.ContainsAny(string(raw[0]), seps) {
		return nil, fmt.Errorf("语法错误: CML禁止以关系符开头:'%c'", raw[0])
	}
	if strings.ContainsAny(string(raw[n-1]), seps) {
		return nil, fmt.Errorf("语法错误: CML禁止以关系符开头: '%c'", raw[n-1])
	}

	/**
	这里使用简易有限状态机解析分隔符和Token，按字节读取，不是字符哦。线性扫码比正则匹配性能远高
	1、对于 A 和 C 模式，Token 是全量编码过的
	2、对于 Q 和 P 模式，Token 可能是明文或带 ! 的编码文
	*/
	var fragments CmlFragments
	lastIdx := 0
	for i := 0; i < len(raw); i++ {
		/**
		因为语法规定编码或原文token，不可以包含特殊字符，且多字节字符的每一段字节编码开头都是1
		所以只要出现分割符字节就一定是分隔符。
		*/
		b := raw[i]

		switch b {
		case '@', '.', '+', ':', ' ': // byte（uint8）与字符rune（uint32）可直接隐式对比，性能最佳
			// 1. 提取 Token
			tokenPart := raw[lastIdx:i]

			// --- 严格校验逻辑 ---
			if tokenPart == "" {
				return nil, fmt.Errorf("非法CML: 空token在字节位置处: %d", i)
			}
			// 正常解码Token
			val, err := decodeToken(tokenPart, mode)
			if err != nil {
				return nil, err
			}
			//将token加入基元序列
			fragments.Tokens = append(fragments.Tokens, val)

			// 2. 提取分隔符，假如基元序列
			fragments.Relations = append(fragments.Relations, string(b))

			// 3. 更新扫描索引
			lastIdx = i + 1
		default:
			// 继续扫描下一个字节
		}
	}
	/**
	剩余的一个必然是不包含关系符的Token，可能包含转义叹号
	由于前置校验过首尾必然不是分隔符，所以最后的token也必然不是空串
	*/
	finalToken := raw[lastIdx:]
	//具体解码见分晓
	val, err := decodeToken(finalToken, mode)
	if err != nil {
		return nil, err
	}
	//将最后一个token加入基元序列
	fragments.Tokens = append(fragments.Tokens, val)
	return &fragments, nil
}
