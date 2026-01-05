package cml

/**
--- 一、基元序列检查和编码 ---
*/

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/shengdoushi/base58"
)

// 注意：由于Go语言限制，[]*CmlElement 不能直接作为接收者挂载方法
// 所以需要使用别名抽象，或者改成命令函数来实现

func IsCmlElements(elements []*CmlElement) error {
	if len(elements) == 0 {
		return errors.New("CML基元序列不应为空")
	}
	if len(elements)%2 == 0 {
		return errors.New("CML基元序列组成数一定是奇数!")
	}
	// CML物理优势：Token与分隔符严格交替
	for i, el := range elements {
		expected := TypeToken
		if i%2 != 0 {
			expected = TypeSeparator
		}
		if el.Type != expected {
			return fmt.Errorf("序列错误在索引 %d: 不是期望的基元类型 %d", i, expected)
		}
	}
	return nil
}

/**
------------------------编码类方法--------------------------
*/

// EncodeA 编码成 a 模式，双层base58，字符集普适性最好，但性能不利于大规模场景
func EncodeA(elements []*CmlElement) (string, error) {
	var sb strings.Builder //使用自动扩容的切片来避免循环分配，提升性能
	for _, el := range elements {
		if el.Type == TypeToken {
			//token级编码
			sb.WriteString(base58.Encode([]byte(el.Value), base58.BitcoinAlphabet))
		} else {
			sb.WriteString(el.Value)
		}
	}
	//整体编码
	payload := base58.Encode([]byte(sb.String()), base58.BitcoinAlphabet)
	return "a" + payload, nil
}

// EncodeP 编码成 c 模式（高性能）
func EncodeC(e []*CmlElement) (string, error) {
	var sb strings.Builder
	for _, el := range e {
		if el.Type == TypeToken {
			sb.WriteString(base64.RawURLEncoding.EncodeToString([]byte(el.Value)))
		} else {
			sb.WriteString(el.Value)
		}
	}
	return "c" + base64.RawURLEncoding.EncodeToString([]byte(sb.String())), nil
}

// EncodeP 编码成 p 模式（单层明文混编，最小熵增）
func EncodeP(e []*CmlElement) (string, error) {
	return "p" + buildBase64URLPayload(e), nil
}

// EncodeP 编码成 q 模式（双层混编，在不可读的前提上，提供最小熵增）
func EncodeQ(e []*CmlElement) (string, error) {
	payload := buildBase64URLPayload(e)
	return "q" + base64.RawURLEncoding.EncodeToString([]byte(payload)), nil
}

// --- base64URL是通用逻辑，除了A模式之外，CPQ都需要 ---
func buildBase64URLPayload(e []*CmlElement) string {
	var sb strings.Builder
	seps := "@.+: !"
	for _, el := range e {
		if el.Type == TypeToken {
			needsEncode := false
			for _, s := range seps {
				if strings.Contains(el.Value, string(s)) {
					needsEncode = true
					break
				}
			}
			if needsEncode {
				sb.WriteString(base64.RawURLEncoding.EncodeToString([]byte(el.Value)) + "!")
			} else {
				sb.WriteString(el.Value)
			}
		} else {
			sb.WriteString(el.Value)
		}
	}
	return sb.String()
}
