package internal

/**
--- 一、CML不同模式之间，端到端的编解码互转 ---
*/
import (
	"errors"
	"fmt"
)

/*
*  检查CML编码是否合法
 */
func IsCML(encoded string) error {
	// 进一步检查是否可以正常两层解码成基元序列
	_, err := CML2Elements(encoded)
	if err != nil {
		return err
	}
	// 最后检查基元是否合法
	return nil
}

/*
*
-----------------简单验证字符串是否为合法的CML编码格式---------------
*/
func cmlBaseCheck(encoded string) error {
	// 长度
	if len(encoded) < 2 {
		return errors.New("CML长度非法")
	}
	// 模式标识
	mode := encoded[0]
	if mode != 'a' && mode != 'c' && mode != 'q' && mode != 'p' {
		return fmt.Errorf("不支持的CML编码模式: %c", mode)
	}
	return nil
}

/**
------------------------编码类方法--------------------------
*/

// 将cml编码，转换成a模式的cml
func CML2A(encoded string) (string, error) {
	//先解析出基元序列，内部会检查
	elements, err := CML2Fragments(encoded) //性能更高
	//elements, err := CML2Elements(encoded) //单序列性能弱一点
	if err != nil {
		return "", err
	}
	return elements.EncodeA()
}

// 将cml编码，转换成c模式的cml
func CML2C(encoded string) (string, error) {
	elements, err := CML2Fragments(encoded)
	if err != nil {
		return "", err
	}
	return elements.EncodeC()
}

// 将cml编码，转换成p模式的cml
func CML2P(encoded string) (string, error) {
	elements, err := CML2Fragments(encoded)
	if err != nil {
		return "", err
	}
	return elements.EncodeP()
}

// 将cml编码，转换成q模式的cml
func CML2Q(encoded string) (string, error) {
	elements, err := CML2Fragments(encoded)
	if err != nil {
		return "", err
	}
	return elements.EncodeQ()
}
