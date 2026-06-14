package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// P0-安全: 写入文件权限应为 0600（仅所有者可读写）
func TestWriteFileTool_FilePermissions(t *testing.T) {
	// Windows 忽略权限检查
	if runtime.GOOS == "windows" {
		t.Skip("skipping file permissions test on Windows")
	}

	tmpDir := t.TempDir()
	tool := NewWriteFileTool(tmpDir)

	args := writeFileArgs{
		Path:    "test.txt",
		Content: "hello world",
	}
	rawArgs, _ := json.Marshal(args)

	_, err := tool.Execute(context.Background(), rawArgs)
	if err != nil {
		t.Fatalf("WriteFileTool failed: %v", err)
	}

	testFile := filepath.Join(tmpDir, "test.txt")
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

// P0-安全: 自动创建的父目录权限应为 0750 或更严格
func TestWriteFileTool_DirPermissions(t *testing.T) {
	// Windows 忽略权限检查
	if runtime.GOOS == "windows" {
		t.Skip("skipping directory permissions test on Windows")
	}

	tmpDir := t.TempDir()
	tool := NewWriteFileTool(tmpDir)

	args := writeFileArgs{
		Path:    "subdir/nested/test.txt",
		Content: "hello",
	}
	rawArgs, _ := json.Marshal(args)

	_, err := tool.Execute(context.Background(), rawArgs)
	if err != nil {
		t.Fatalf("WriteFileTool failed: %v", err)
	}

	subDir := filepath.Join(tmpDir, "subdir")
	info, err := os.Stat(subDir)
	if err != nil {
		t.Fatalf("failed to stat subdir: %v", err)
	}

	perm := info.Mode().Perm()
	// 目录权限应该是 0750 或更严格
	if perm&0022 != 0 {
		t.Errorf("dir %s has permissions %o, should be 0750 or stricter", subDir, perm)
	}
}
