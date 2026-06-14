package tools

import (
	"strings"
	"testing"
)

// P1-01: L1 重叠计数 Bug — "aaa" 中找 "aa" 应匹配1处而非2处
func TestFuzzyReplace_L1_OverlappingCount(t *testing.T) {
	content := `aaa`
	oldText := `aa`
	newText := `bb`

	result, err := fuzzyReplace(content, oldText, newText)
	if err != nil {
		t.Fatalf("重叠计数 bug: L1 应匹配1处，算法却报错或降级: %v", err)
	}
	if result != "bba" {
		t.Fatalf("expected 'bba', got: %s", result)
	}
}

// P1-02: L3 替换后必须保留原始缩进
func TestFuzzyReplace_L3_PreserveIndent(t *testing.T) {
	content := `func main() {
    println("hello")
}`
	oldText := `println("hello")`
	newText := `println("world")`

	result, err := fuzzyReplace(content, oldText, newText)
	if err != nil {
		t.Fatalf("L3 匹配失败: %v", err)
	}
	// 验证：替换后仍然保留了原始的4空格缩进
	if !strings.Contains(result, `    println("world")`) {
		t.Fatalf("L3 替换后缩进丢失，结果:\n%s", result)
	}
}

// P1-03: Windows \r\n 文件（content 有 \r）L4 应正常匹配
func TestFuzzyReplace_L4_WindowsLineEnding(t *testing.T) {
	// 模拟 Windows 文件：\r\n 换行
	content := "func main() {\r\n    println(\"hello\")\r\n}"
	oldText := `println("hello")`
	newText := `println("world")`

	result, err := fuzzyReplace(content, oldText, newText)
	if err != nil {
		t.Fatalf("L4 Windows 换行符匹配失败: %v", err)
	}
	if !strings.Contains(result, `println("world")`) {
		t.Fatalf("L4 替换失败，结果:\n%s", result)
	}
}

// P1-04: L4 多行 newText 应正确拆分替换
func TestFuzzyReplace_L4_MultiLineNewText(t *testing.T) {
	content := `func main() {
    println("start")
}`
	oldText := `println("start")`
	newText := `println("middle")
    println("end")`

	result, err := fuzzyReplace(content, oldText, newText)
	if err != nil {
		t.Fatalf("L4 多行替换失败: %v", err)
	}
	if !strings.Contains(result, `println("middle")`) {
		t.Fatalf("多行替换后未找到 middle，结果:\n%s", result)
	}
	if !strings.Contains(result, `println("end")`) {
		t.Fatalf("多行替换后未找到 end，结果:\n%s", result)
	}
}

// P1-05: L2 换行符归一化后精确匹配
func TestFuzzyReplace_L2_NormalizeCRLF(t *testing.T) {
	content := "func main() {\r\n    println(\"hello\")\r\n}"
	oldText := "println(\"hello\")"
	newText := `println("world")`

	result, err := fuzzyReplace(content, oldText, newText)
	if err != nil {
		t.Fatalf("L2 CRLF 归一化匹配失败: %v", err)
	}
	if !strings.Contains(result, `println("world")`) {
		t.Fatalf("L2 替换失败，结果:\n%s", result)
	}
}

// P1-06: L4 多行 oldText 逐行去缩进匹配
func TestFuzzyReplace_L4_MultiLineOldText(t *testing.T) {
	content := `func main() {
    a := 1
    b := 2
}`
	// 大模型给的 oldText 没有缩进
	oldText := `a := 1
b := 2`
	newText := `a := 100
b := 200`

	result, err := fuzzyReplace(content, oldText, newText)
	if err != nil {
		t.Fatalf("L4 多行匹配失败: %v", err)
	}
	if !strings.Contains(result, `a := 100`) || !strings.Contains(result, `b := 200`) {
		t.Fatalf("L4 多行替换失败，结果:\n%s", result)
	}
}