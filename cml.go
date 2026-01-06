package cml

/**
--- 一、CML不同模式之间，端到端的编解码互转 ---
*/
import (
	"cml/internal"
)

/*
*  检查CML编码是否合法
 */
func IsCML(encoded string) error {
	return internal.IsCML(encoded)
}

/**
------------------------编码类方法--------------------------
*/
// 将cml编码，转换成a模式的cml
func CML2A(encoded string) (string, error) {
	return internal.CML2A(encoded)
}

// 将cml编码，转换成c模式的cml
func CML2C(encoded string) (string, error) {
	return internal.CML2C(encoded)
}

// 将cml编码，转换成p模式的cml
func CML2P(encoded string) (string, error) {
	return internal.CML2P(encoded)
}

// 将cml编码，转换成q模式的cml
func CML2Q(encoded string) (string, error) {
	return internal.CML2Q(encoded)
}

// CML2Elements 将编码后的CML字符串解析为基元序列
func CML2Elements(encoded string) (CMLSingle, error) {
	return internal.CML2Elements(encoded)
}

/*
-----------------------------定义完全等价的别名-----------------------------
---------------不需要强制转换到底层类型的指针类型，零复制、零开销--------------
*/

/*
这是两种中间结果实现之一：基元类型抽象的单序列
<token>，<separator>，<token>，<separator>，...<token>
*/
type CMLSingle = internal.CmlElements

/*
这是两种中间结构实现之之二：基元类型的分类双序列
奇数序列<token>，<token>，...<token>
偶数序列<separator>，<separator>，...<separator>
由于CML有交替规律，使用双列表，比使用基元类型抽象的编解码转换性能更高，可以显著减少内存分配次数
*/
type CMLDouble = internal.CmlFragments
