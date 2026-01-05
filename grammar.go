package cml

/**
------------------语法规范-----------------------
*/

const (
	// 关系分隔符定义
	SepRemarkAt     = "@" // 补充关系
	SepLineDot      = "." // 线性递进
	SepSetXX        = "+" // 并列集合
	SepMapping      = ":" // 映射关系
	SepCombineSpace = " " // 组合关系
)

// 编码模式
const (
	ModeA = 'a' // Double Base58
	ModeC = 'c' // Double Base64URL
	ModeQ = 'q' // Hybrid + Global Base64URL
	ModeP = 'p' // Hybrid Plaintext

	//编码标识符，PQ模式专用
	EncodedExclaim = "!" //混编转义符，标明这是一个被强制编码的非原文token
)

/**
------------------实现规范-----------------------
由于有交替规律，使用双列表，比使用基元类型抽象性能更高
*/

// 基元类型
type CmlElementType int

const (
	TypeToken CmlElementType = iota
	TypeSeparator
)

//语义基元，构建序列的单元
type CmlElement struct {
	Type  CmlElementType
	Value string
}

//解析后的两类基元的序列
type CmlFragments struct {
	Tokens    []string // 所有的实体内容
	Relations []string // 所有的关系符 (@, ., +, :, 空格)
}
