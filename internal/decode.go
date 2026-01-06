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

// 上层整体解码,返回编码模式和有效荷载
func decodePayload(encoded string) (uint8, string, error) {
	//基础的长度和首字符检查
	if err := cmlBaseCheck(encoded); err != nil {
		return 0, "", err
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
			return 0, "", fmt.Errorf("base58整体解码失败: %v", err)
		}
		rawPayload = string(b)

	case ModeC, ModeQ:
		b, err := base64.RawURLEncoding.DecodeString(payload)
		if err != nil {
			return 0, "", fmt.Errorf("base64URL整体解码失败: %v", err)
		}
		rawPayload = string(b)
	case ModeP:
		//p模式是单层架构
		rawPayload = payload
	}
	//第二层荷载解码
	return mode, rawPayload, nil
}

// token级解码原文
func decodeToken(rawToken string, mode uint8) (string, error) {
	switch mode {
	case ModeA:
		// 直接解码原文
		b, err := base58.Decode(rawToken, base58.BitcoinAlphabet)
		if err != nil {
			return "", fmt.Errorf("base58解码token原文失败: %v", err)
		}
		return string(b), nil
	case ModeC:
		// 语言强制要求原文编码是不能带有等号的原始格式，不需要适配=！
		b, err := base64.RawURLEncoding.DecodeString(rawToken)
		if err != nil {
			return "", fmt.Errorf("base64URL解码token原文失败: %v", err)
		}
		return string(b), nil
	case ModeQ, ModeP:
		// 检查是否有混编转义符 '!'，有则切除，进行解码
		if before, ok := strings.CutSuffix(rawToken, "!"); ok {
			// 只有不带 '!' 的原荷载部分才进行 Base64 解码
			b, err := base64.RawURLEncoding.DecodeString(before)
			if err != nil {
				// 如果解码失败，说明数据在传输中可能损坏
				return "", fmt.Errorf("base64URL解码混编token原文失败 [%s]: %w", rawToken, err)
			}
			return string(b), nil
		}

		// 是明文就直接返回
		return rawToken, nil
	}
	return rawToken, nil
}
