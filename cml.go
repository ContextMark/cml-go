// CML是语义时代的Markdown。
// 它的目标是让兼具人类可读和机器可运算特征的关系结构片段，成为**可大规模计算**、**任意传输**、**分布式存储**的语义中间层。
package cml

/**
本 SDK 代码采用 Apache 2.0 协议，但其解析的 CML 语法标准遵循 CML 双许可协议。
1、商业平台类用途请参阅 CML 许可声明。
2、禁止演绎核心语法规范，需对各种模式的编解码规则兼容。
*/

import (
	"github.com/ContextMark/cml-go/internal"
)

/*
检查CML编码是否合法
*/
func IsCML(encoded string) error {
	return internal.IsCML(encoded)
}

/**
------------------------编码模式互换类方法--------------------------
*/

// CML2A 将cml编码，转换成a模式的cml编码
func CML2A(encoded string) (string, error) {
	return internal.CML2A(encoded)
}

// CML2C 将cml编码，转换成c模式的cml编码
func CML2C(encoded string) (string, error) {
	return internal.CML2C(encoded)
}

// CML2P 将cml编码，转换成p模式的cml编码
func CML2P(encoded string) (string, error) {
	return internal.CML2P(encoded)
}

// CML2Q 将cml编码，转换成q模式的cml编码
func CML2Q(encoded string) (string, error) {
	return internal.CML2Q(encoded)
}

/**
------------------------转换类方法--------------------------
*/

// CML2Elements 将CML字符串解析为类型抽象的单基元序列
func CML2Elements(encoded string) (*CMLSingle, error) {
	return internal.CML2Elements(encoded)
}

// CML2Fragments 将CML字符串解析为基于类型分类的双基元序列
func CML2Fragments(encoded string) (*CMLDouble, error) {
	return internal.CML2Fragments(encoded)
}

// 将cml编码转换成md反引号格式
func ToMarkdown(encoded string) string {
	return toMarkdown(encoded)
}

// 将反引号编码的md格式转换成基元序列
func FromMarkdown(md string) ([]string, error) {
	return fromMarkdown(md)
}

// New 手动构造CML双基元序列
func New(arr []string) (*CMLDouble, error) {
	return internal.New(arr)
}

/*
-----------------------------定义完全等价的别名-----------------------------
---------------不需要强制转换到底层类型的指针类型，零复制、零开销--------------
*/

// 中间结构实现之一：基元类型抽象的单序列
// <token>，<separator>，<token>，<separator>，...<token>
type CMLSingle = internal.CmlElements

// 中间结构实现之二：基元类型的分类双序列
// - 奇数序列<token>，<token>，...<token>
// - 偶数序列<separator>，<separator>，...<separator>
// 由于CML有交替规律，使用双列表，比使用基元类型抽象的编解码转换性能更高，可以显著减少内存分配次数
type CMLDouble = internal.CmlFragments
