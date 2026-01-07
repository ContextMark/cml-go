package internal

/**
--- 二、基元和CML字符串互转 ---
*/
import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/shengdoushi/base58"
)

// CML2Elements 将CML字符串解析为基元序列
func CML2Elements(encoded string) (*CmlElements, error) {
	//基础的长度和首字符检查
	if err := cmlBaseCheck(encoded); err != nil {
		return nil, err
	}

	mode := encoded[0]
	payload := encoded[1:]
	var rawPayload string

	// 第一层整体解码：还原语义荷载
	switch mode {
	case ModeA:
		// 解码整体荷载
		b, err := base58.Decode(payload, base58.BitcoinAlphabet)
		if err != nil {
			return nil, fmt.Errorf("base58整体解码失败: %v", err)
		}
		rawPayload = string(b)

	case ModeC, ModeQ:
		b, err := base64.RawURLEncoding.DecodeString(payload)
		if err != nil {
			return nil, fmt.Errorf("base64URL整体解码失败: %v", err)
		}
		rawPayload = string(b)
	case ModeP:
		//p模式是单层架构
		rawPayload = payload
	}
	//将实际的荷载解码成单序列
	return parseRawPayload2Single(rawPayload, mode)
}

// parseRawPayload 解析解密后的原始荷载字符串
func parseRawPayload2Single(raw string, mode uint8) (*CmlElements, error) {
	n := len(raw)
	if n == 0 {
		return nil, fmt.Errorf("非法的空CML")
	}
	// 定义5个分隔符集合
	const seps = "@.+: "
	//关系符必然不出现在首位和尾位
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
	var elements CmlElements
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
			elements = append(elements, &CmlElement{Type: TypeToken, Value: val})

			// 2. 提取分隔符，假如基元序列
			elements = append(elements, &CmlElement{Type: TypeSeparator, Value: string(b)})

			// 3. 更新扫描索引
			lastIdx = i + 1
		default:
			// 继续扫描下一个字节
		}
	}
	// 剩余的一个必然是Token
	finalToken := raw[lastIdx:]
	val, err := decodeToken(finalToken, mode)
	if err != nil {
		return nil, err
	}
	//将token加入基元序列
	elements = append(elements, &CmlElement{Type: TypeToken, Value: val})
	return &elements, nil
}
