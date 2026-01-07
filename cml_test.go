package cml_test

import (
	"testing"

	"github.com/ContextMark/cml-go"

	/**
	require和assert是互补的走左还是走右的问题：
	1、require当断言失败时，子测试“停下来”，对于全流程前后依赖测试可以避免同一个bug报错无数次。
	2、assert，当断言失败时，子测试“继续走”。
	*/
	// 当断言失败时，子测试“继续走”。
	"github.com/stretchr/testify/require"
	// "github.com/stretchr/testify/assert"
)

/**
go test -v ./...    //测试所有package，同时-v打印没有给case测试状态
go test -v          //这里只需要测试子包，接口都封装到了外层
go test             //最简洁的，只有没通过的case才需要详情
*/

// 核心语法测试
func TestCML_CorePipeline(t *testing.T) {
	// 定义测试数据集
	tests := []struct {
		name      string //case名
		expectErr bool   //预期
		//不同测试方向，内部需要适配
		input  []string //手动输入构造要素
		rawCML string   //或者构造编码结果
	}{
		{
			name: "正常场景：标准全流程验证",
			input: []string{
				"万有引力",
				"@",
				"1867年《自然哲学的数学原理》",
				"@",
				"牛顿",
			},
			expectErr: false,
		},
		{
			name:      "正常场景：token带反引号和分隔符",
			input:     []string{"haha`Key", "@", "hehe`@Value!"},
			expectErr: false,
		},
		{
			name: "错误场景❌：非法分隔符",
			input: []string{
				"万有引力",
				"xx",
				"《自然哲学的数学原理》",
			},
			expectErr: true,
		},
		{
			name: "错误场景：不满足关系符不在首尾原则",
			input: []string{
				"万有引力",
				"@",
			},
			expectErr: true,
		},
		{
			name:      "错误场景：非法字符串直接解析",
			rawCML:    "INVALID_FORMAT_DATA",
			expectErr: true,
		},
		{
			name:      "错误场景：空输入校验",
			input:     nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 使用 testify 的 assert 对象或require对象，它会自动处理 t.Errorf 并让代码更整洁
			// ast := assert.New(t)
			ast := require.New(t)

			// --- 逻辑分支 A：测试原始 CML 字符串解析 ---
			if tt.rawCML != "" {
				err := cml.IsCML(tt.rawCML)
				if tt.expectErr {
					ast.Error(err) //应该返回错误，没有错误就算测试失败
				} else {
					ast.NoError(err) //不应该返回错误，反之有错误就算测试失败
				}
				return
			}

			// --- 逻辑分支 B：全流程测试 ---

			// 1. 创建与编码
			cDouble, err := cml.New(tt.input)
			if tt.expectErr {
				ast.Error(err, "初始化阶段应报错")
				return
			}
			ast.NoError(err)

			pStr, err := cDouble.EncodeP()
			ast.NoError(err)
			ast.NoError(cml.IsCML(pStr))

			// 2. 模式轮转：P -> Q -> A -> C
			qStr, err := cml.CML2Q(pStr)
			ast.NoError(err)

			aStr, err := cml.CML2A(qStr)
			ast.NoError(err)

			cStr, err := cml.CML2C(aStr)
			ast.NoError(err)

			// 3. 结构化解析验证
			elements, err := cml.CML2Elements(cStr)
			ast.NoError(err)
			ast.NoError(elements.IsValid())

			fragments, err := cml.CML2Fragments(aStr)
			ast.NoError(err)
			ast.NoError(fragments.IsValid())

			// 4. Markdown 往返一致性
			md := cml.ToMarkdown(pStr)
			recovered, err := cml.FromMarkdown(md)
			ast.NoError(err)
			ast.ElementsMatch(tt.input, recovered, "Markdown 还原内容不匹配")

			// 5. 闭环验证：C -> P
			finalP, err := cml.CML2P(cStr)
			ast.NoError(err)
			ast.Equal(pStr, finalP, "闭环转换 P 模式不一致")
		})
	}
}

// 上层markdown测试
func TestCML_Markdown(t *testing.T) {
	ast := require.New(t)

	// 1. 专项测试用例：覆盖特殊边界情况
	mdTests := []struct {
		name     string
		input    []string
		expected string // 期望生成的 Markdown 格式
	}{
		{
			name:     "Markdown：普通文本测试",
			input:    []string{"语言名", "@", "cml"},
			expected: "`语言名` @ `cml`",
		},
		{
			name:  "Markdown：带反引号的文本转义",
			input: []string{"Code`Ref", "@", "Value"},
			// 假设你的 toMarkdown 内部处理了反引号转义，通常是双反引号或空格处理
			expected: "`` Code`Ref `` @ `Value`",
		},
	}

	for _, tt := range mdTests {
		t.Run(tt.name, func(t *testing.T) {
			// 先通过 New 构造出 CML 编码串（模拟真实链路）
			c, err := cml.New(tt.input)
			ast.NoError(err)
			pStr, _ := c.EncodeP()

			// 测试 ToMarkdown
			gotMd := cml.ToMarkdown(pStr)
			t.Logf("生成md: %s", gotMd) //t.Logf只有当测试 Case 失败（Fail）时，它打印的内容才会显示出来。
			// 这里可以根据你具体的 Markdown 格式规范校验 tt.expected

			// 测试 FromMarkdown (核心：还原能力)
			recovered, err := cml.FromMarkdown(gotMd)
			ast.NoError(err)
			ast.Equal(tt.input, recovered, "Markdown 还原后的切片应与原始输入完全一致")
		})
	}

	// 2. 异常接口测试
	t.Run("Markdown：非法输入测试", func(t *testing.T) {
		// 测试空字符串
		res, err := cml.FromMarkdown("")
		ast.Error(err, "空字符串解析应返回错误")
		ast.Nil(res)

		// 测试格式错误的 Markdown (如反引号不闭合)
		res, err = cml.FromMarkdown("`unclosed backtick")
		ast.Error(err)
	})
}
