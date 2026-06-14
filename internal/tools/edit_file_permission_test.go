package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// P0-安全: 编辑后的文件权限应为 0600
func TestEditFileTool_FilePermissions(t *testing.T) {
	// Windows 忽略权限检查
	if runtime.GOOS == "windows" {
		t.Skip("skipping file permissions test on Windows")
	}

	tmpDir := t.TempDir()
	tool := NewEditFileTool(tmpDir)

	// 先创建文件
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	args := editFileArgs{
		Path:    "test.txt",
		OldText: "hello",
		NewText: "goodbye",
	}
	rawArgs, _ := json.Marshal(args)

	_, err := tool.Execute(context.Background(), rawArgs)
	if err != nil {
		t.Fatalf("EditFileTool failed: %v", err)
	}

	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("failed to stat test file: %v", err)
	}

	perm := info.Mode().Perm()
	// 文件权限应该是 0600 或更严格
	if perm&0066 != 0 {
		t.Errorf("file %s has permissions %o, should be 0600 or stricter", testFile, perm)
	}
}