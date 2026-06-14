package observability

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// P0-安全: 追踪文件写入权限应为 0600（仅所有者可读写）
func TestExportTraceToFile_Permissions(t *testing.T) {
	tmpDir := t.TempDir()

	rootSpan := &Span{
		Name:       "test-span",
		StartTime:  time.Now(),
		EndTime:    time.Now().Add(100 * time.Millisecond),
		Attributes: map[string]interface{}{"test": "attr"},
		Children:   nil,
	}

	sessionID := "test-session"
	err := ExportTraceToFile(rootSpan, tmpDir, sessionID)
	if err != nil {
		t.Fatalf("ExportTraceToFile failed: %v", err)
	}

	// 检查生成的追踪文件权限
	tracesDir := filepath.Join(tmpDir, ".claw", "traces")
	files, err := os.ReadDir(tracesDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	for _, f := range files {
		info, err := f.Info()
		if err != nil {
			t.Fatalf("failed to get file info: %v", err)
		}
		perm := info.Mode().Perm()
		// 追踪文件应该是 0600 或更严格（所有者读写，无组/其他权限）
		// Windows 下忽略此检查
		if runtime.GOOS != "windows" && perm&0066 != 0 {
			t.Errorf("trace file %s has permissions %o, should be 0600 or stricter", f.Name(), perm)
		}
	}
}

// P0-安全: 追踪目录权限应为 0750 或更严格
func TestTraceDir_Permissions(t *testing.T) {
	// Windows 忽略目录权限测试
	if runtime.GOOS == "windows" {
		t.Skip("skipping directory permissions test on Windows")
	}

	tmpDir := t.TempDir()

	rootSpan := &Span{
		Name:       "test",
		StartTime:  time.Now(),
		EndTime:    time.Now(),
		Attributes: map[string]interface{}{},
	}
	ExportTraceToFile(rootSpan, tmpDir, "s1")
	ExportTraceToFile(rootSpan, tmpDir, "s2")

	traceDir := filepath.Join(tmpDir, ".claw", "traces")
	info, err := os.Stat(traceDir)
	if err != nil {
		t.Fatalf("trace dir should exist: %v", err)
	}

	perm := info.Mode().Perm()
	// 目录应该是 0750 或更严格（无组/其他写权限）
	if perm&0022 != 0 {
		t.Errorf("trace dir has permissions %o, should be 0750 or stricter", perm)
	}
}