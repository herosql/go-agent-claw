// internal/tools/read_file_bench_test.go
package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// BenchmarkReadFileConcurrent 基准测试：验证并发读取同一文件的安全性
func BenchmarkReadFileConcurrent(b *testing.B) {
	// 1. 准备测试文件
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "nginx.conf")
	testContent := "server {\n    listen 80;\n    server_name localhost;\n}\n"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		b.Fatal(err)
	}

	tool := NewReadFileTool(tmpDir)
	args, _ := json.Marshal(map[string]string{"path": "nginx.conf"})
	ctx := context.Background()

	// 2. 并发压测：模拟 loop.go 中多个 goroutine 并发调用同一工具
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := tool.Execute(ctx, args)
			if err != nil {
				b.Errorf("并发读取失败: %v", err)
			}
		}
	})
}

// BenchmarkReadFileDifferentFiles 基准测试：并发读取不同文件
func BenchmarkReadFileDifferentFiles(b *testing.B) {
	tmpDir := b.TempDir()
	fileCount := 10
	files := make([]string, fileCount)
	for i := 0; i < fileCount; i++ {
		files[i] = filepath.Join(tmpDir, "file_"+string(rune('a'+i))+".txt")
		if err := os.WriteFile(files[i], []byte("content"), 0644); err != nil {
			b.Fatal(err)
		}
	}

	tool := NewReadFileTool(tmpDir)
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			path := "file_" + string(rune('a'+i%fileCount)) + ".txt"
			args, _ := json.Marshal(map[string]string{"path": path})
			_, err := tool.Execute(ctx, args)
			if err != nil {
				b.Errorf("读取失败: %v", err)
			}
			i++
		}
	})
}

// TestReadFileConcurrent_Sequential 对比测试：顺序读取（用于验证并发问题）
func TestReadFileConcurrent_Sequential(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "nginx.conf")
	testContent := "server { listen 80; }"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewReadFileTool(tmpDir)
	args, _ := json.Marshal(map[string]string{"path": "nginx.conf"})
	ctx := context.Background()

	// 顺序执行 100 次，验证基本功能
	for i := 0; i < 100; i++ {
		_, err := tool.Execute(ctx, args)
		if err != nil {
			t.Errorf("第 %d 次读取失败: %v", i, err)
		}
	}
}

// TestReadFileConcurrent_Parallel 并发测试：暴露竞态条件
func TestReadFileConcurrent_Parallel(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "nginx.conf")
	testContent := "server { listen 80; }"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewReadFileTool(tmpDir)
	args, _ := json.Marshal(map[string]string{"path": "nginx.conf"})
	ctx := context.Background()

	var wg sync.WaitGroup
	errCh := make(chan error, 100)

	// 并发执行 100 次
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := tool.Execute(ctx, args)
			if err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("并发读取出现错误: %v", err)
	}
}
