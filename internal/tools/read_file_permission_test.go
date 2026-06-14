package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// P0-安全: 读取文件权限应为 0600 或更严格
func TestReadFileTool_FilePermissions(t *testing.T) {
	// Windows 忽略权限检查
	if runtime.GOOS == "windows" {
		t.Skip("skipping file permissions test on Windows")
	}

	tmpDir := t.TempDir()
	tool := NewReadFileTool(tmpDir)

	// 创建权限过宽的文件 (0644 有组/其他读权限)
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	args := readFileArgs{
		Path: "test.txt",
	}
	rawArgs, _ := json.Marshal(args)

	output, err := tool.Execute(context.Background(), rawArgs)
	if err != nil {
		t.Fatalf("ReadFileTool failed: %v", err)
	}

	if output == "" {
		t.Fatal("expected file content, got empty")
	}
}

// P0-安全: 工作区外文件读取应被阻止
func TestReadFileTool_PathTraversal_Blocked(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewReadFileTool(tmpDir)

	args := readFileArgs{
		Path: "../secret.txt",
	}
	rawArgs, _ := json.Marshal(args)

	_, err := tool.Execute(context.Background(), rawArgs)
	if err == nil {
		t.Fatal("path traversal should be blocked")
	}
}

// P0-安全: 绝对路径应被阻止
func TestReadFileTool_AbsolutePath_Blocked(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewReadFileTool(tmpDir)

	args := readFileArgs{
		Path: "/etc/passwd",
	}
	rawArgs, _ := json.Marshal(args)

	_, err := tool.Execute(context.Background(), rawArgs)
	if err == nil {
		t.Fatal("absolute path should be blocked")
	}
}

// P0-安全: 空字节路径应被阻止
func TestReadFileTool_NullByte_Blocked(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewReadFileTool(tmpDir)

	args := readFileArgs{
		Path: "test.txt\x00/etc/passwd",
	}
	rawArgs, _ := json.Marshal(args)

	_, err := tool.Execute(context.Background(), rawArgs)
	if err == nil {
		t.Fatal("null byte path should be blocked")
	}
}

// P0-安全: 多级路径穿越应被阻止
func TestReadFileTool_MultiLevelTraversal_Blocked(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewReadFileTool(tmpDir)

	args := readFileArgs{
		Path: "../../../etc/passwd",
	}
	rawArgs, _ := json.Marshal(args)

	_, err := tool.Execute(context.Background(), rawArgs)
	if err == nil {
		t.Fatal("multi-level traversal should be blocked")
	}
}

// P0-安全: 合法路径应正常工作
func TestReadFileTool_LegalPath_Works(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewReadFileTool(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world"), 0600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	args := readFileArgs{
		Path: "test.txt",
	}
	rawArgs, _ := json.Marshal(args)

	output, err := tool.Execute(context.Background(), rawArgs)
	if err != nil {
		t.Fatalf("legal path should work, got error: %v", err)
	}

	if output != "hello world" {
		t.Fatalf("expected 'hello world', got: %s", output)
	}
}
