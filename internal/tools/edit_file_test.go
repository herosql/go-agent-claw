package tools

import (
	"strings"
	"testing"
)

// P0-04: L1 精确匹配，唯一时正确替换
func TestFuzzyReplace_L1_UniqueMatch(t *testing.T) {
	content := `package main

func main() {
	println("hello")
}`
	oldText := `println("hello")`
	newText := `println("world")`

	result, err := fuzzyReplace(content, oldText, newText)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, `println("world")`) {
		t.Fatalf("expected new text not found in result:\n%s", result)
	}
	if strings.Contains(result, `println("hello")`) {
		t.Fatalf("old text should be replaced, still found in result:\n%s", result)
	}
}

// P0-05: L1 匹配 0 处返回 error
func TestFuzzyReplace_L1_NotFound(t *testing.T) {
	content := `func main() { println("hello") }`
	oldText := `println("不存在")`
	newText := `println("world")`

	_, err := fuzzyReplace(content, oldText, newText)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "未找到") {
		t.Fatalf("error message should mention '未找到', got: %v", err)
	}
}

// P0-06: L1 匹配 >1 处返回 error
func TestFuzzyReplace_L1_MultipleMatches(t *testing.T) {
	content := `func foo() { x = 1 }
func bar() { x = 1 }`
	oldText := `x = 1`
	newText := `x = 2`

	_, err := fuzzyReplace(content, oldText, newText)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "匹配到了") {
		t.Fatalf("error should mention multiple matches or need more context, got: %v", err)
	}
}

// P0-07: L4 逐行去缩进匹配（大模型遗漏缩进仍能匹配）
func TestFuzzyReplace_L4_StripIndent(t *testing.T) {
	// 文件中实际内容有缩进，但 oldText 来自大模型（无缩进）
	content := `func main() {
    println("hello")
}`
	// 大模型给的 oldText 没有缩进
	oldText := `println("hello")`
	newText := `println("world")`

	result, err := fuzzyReplace(content, oldText, newText)
	if err != nil {
		t.Fatalf("L4 fuzzy replace failed: %v", err)
	}
	if !strings.Contains(result, `println("world")`) {
		t.Fatalf("L4 replace failed, result:\n%s", result)
	}
}
