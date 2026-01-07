package internal

/**
--- 一、基元序列检查和编码 ---
由于Go语言限制，[]*CmlElement 不能直接作为类型接收者挂载方法
所以需要别名抽象来实现
*/

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/shengdoushi/base58"
)

/**
--- 序列检查方法 ---
*/

// 检查 CML 基元序列的奇偶交替特征是否合法
func (elements *CmlElements) IsValid() error {

	if elements == nil || len(*elements) == 0 {
		return errors.New("CML基元序列不应为空")
	}
	if len(*elements)%2 == 0 {
		return errors.New("CML基元序列组成数一定是奇数!")
	}
	// CML物理优势：Token与分隔符严格交替
	for i, el := range *elements {
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
func (elements *CmlElements) EncodeA() (string, error) {
	if elements == nil || len(*elements) == 0 {
		return "", errors.New("CML基元序列不应为nil")
	}
	var sb strings.Builder // 使用自动扩容的切片来避免循环分配，提升性能
	for _, el := range *elements {
		if el.Type == TypeToken {
			// token级编码
			sb.WriteString(base58.Encode([]byte(el.Value), base58.BitcoinAlphabet))
		} else {
			sb.WriteString(el.Value)
		}
	}
	// 整体编码
	payload := base58.Encode([]byte(sb.String()), base58.BitcoinAlphabet)
	return "a" + payload, nil
}

// EncodeC 编码成 c 模式（高性能 base64）
func (elements *CmlElements) EncodeC() (string, error) {
	if elements == nil || len(*elements) == 0 {
		return "", errors.New("CML基元序列不应为nil")
	}
	var sb strings.Builder
	for _, el := range *elements {
		if el.Type == TypeToken {
			sb.WriteString(base64.RawURLEncoding.EncodeToString([]byte(el.Value)))
		} else {
			sb.WriteString(el.Value)
		}
	}
	return "c" + base64.RawURLEncoding.EncodeToString([]byte(sb.String())), nil
}

// EncodeP 编码成 p 模式（单层明文混编，最小熵增）
func (elements *CmlElements) EncodeP() (string, error) {
	if elements == nil || len(*elements) == 0 {
		return "", errors.New("CML基元序列不应为nil")
	}
	return "p" + elements.buildMixedPayload(), nil
}

// EncodeQ 编码成 q 模式（双层混编，在不可读的前提上，提供最小熵增）
func (elements *CmlElements) EncodeQ() (string, error) {
	if elements == nil || len(*elements) == 0 {
		return "", errors.New("CML基元序列不应为nil")
	}
	payload := elements.buildMixedPayload()
	return "q" + base64.RawURLEncoding.EncodeToString([]byte(payload)), nil
}

/*
*
--- 通用混编逻辑（PQ模式共用） ---
*/
func (elements *CmlElements) buildMixedPayload() string {
	var sb strings.Builder
	for _, el := range *elements {
		/**
		检查每一个基元:
		1、如果是token类，就要检查是否包含需要规避的保留字符集：5关系符+编码转义符
		2、如果是关系符，直接追加
		*/
		if el.Type == TypeToken {
			// 调用独立函数处理单个 token，决定是采用原值还是编码
			sb.WriteString(processToken(el.Value))
		} else {
			sb.WriteString(el.Value)
		}
	}
	return sb.String()
}
