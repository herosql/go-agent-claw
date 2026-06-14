package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// P0-安全: 路径穿越攻击应被阻止
func TestEditFileTool_PathTraversal_Blocked(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewEditFileTool(tmpDir)

	// 构造路径穿越攻击：../ 穿透到工作区外
	args := editFileArgs{
		Path:    "../secret.txt",
		OldText: "secret",
		NewText: "hacked",
	}
	rawArgs, _ := json.Marshal(args)

	_, err := tool.Execute(context.Background(), rawArgs)

	// 应该返回错误，且明确说明是路径越界
	if err == nil {
		t.Fatal("path traversal attack should be blocked, but got no error")
	}
	if !strings.Contains(err.Error(), "不允许") && !strings.Contains(err.Error(), "禁止") && !strings.Contains(err.Error(), "越界") {
		t.Fatalf("error should explicitly mention path not allowed, got: %v", err)
	}
}

// P0-安全: 合法相对路径应正常工作
func TestEditFileTool_LegalPath_Works(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tool := NewEditFileTool(tmpDir)
	args := editFileArgs{
		Path:    "test.txt",
		OldText: "hello",
		NewText: "goodbye",
	}
	rawArgs, _ := json.Marshal(args)

	output, err := tool.Execute(context.Background(), rawArgs)
	if err != nil {
		t.Fatalf("legal path should work, got error: %v", err)
	}
	if !strings.Contains(output, "成功修改") {
		t.Fatalf("expected success message, got: %s", output)
	}
}

// P0-安全: 绝对路径应被阻止
func TestEditFileTool_AbsolutePath_Blocked(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewEditFileTool(tmpDir)

	args := editFileArgs{
		Path:    "/etc/passwd",
		OldText: "root",
		NewText: "hacked",
	}
	rawArgs, _ := json.Marshal(args)

	_, err := tool.Execute(context.Background(), rawArgs)
	if err == nil {
		t.Fatal("absolute path should be blocked")
	}
	// 绝对路径应该被拒绝（通过 isUserPathSafe 检查）
	if !strings.Contains(err.Error(), "不允许") && !strings.Contains(err.Error(), "禁止") && !strings.Contains(err.Error(), "路径") {
		t.Fatalf("error should mention path restriction, got: %v", err)
	}
}

// P0-安全: 带空字节的路径应被阻止
func TestEditFileTool_NullByte_Blocked(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewEditFileTool(tmpDir)

	args := editFileArgs{
		Path:    "test.txt\x00/etc/passwd",
		OldText: "hello",
		NewText: "world",
	}
	rawArgs, _ := json.Marshal(args)

	_, err := tool.Execute(context.Background(), rawArgs)
	if err == nil {
		t.Fatal("null byte path should be blocked")
	}
}

// P0-安全: 多级路径穿越应被阻止
func TestEditFileTool_MultiLevelTraversal_Blocked(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewEditFileTool(tmpDir)

	args := editFileArgs{
		Path:    "../../../etc/passwd",
		OldText: "root",
		NewText: "hacked",
	}
	rawArgs, _ := json.Marshal(args)

	_, err := tool.Execute(context.Background(), rawArgs)
	if err == nil {
		t.Fatal("multi-level traversal should be blocked")
	}
	if !strings.Contains(err.Error(), "不允许") && !strings.Contains(err.Error(), "禁止") {
		t.Fatalf("error should mention path not allowed, got: %v", err)
	}
}
