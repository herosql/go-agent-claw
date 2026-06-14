// cmd/envcheck/main.go
// 诊断环境变量是否可以被 Go 进程读取
package main

import (
	"fmt"
	"os"
)

func main() {
	envVars := []string{
		"FEISHU_APP_ID",
		"FEISHU_APP_SECRET",
		"FEISHU_ENCRYPT_KEY",
		"FEISHU_VERIFY_TOKEN",
		"ZHIPU_API_KEY",
		"PATH",
	}

	fmt.Println("=== 环境变量诊断 ===")
	for _, key := range envVars {
		val := os.Getenv(key)
		if val == "" {
			fmt.Printf("❌ %s = (空)\n", key)
		} else if key == "PATH" {
			fmt.Printf("✅ %s = [长度 %d 的字符串]\n", key, len(val))
		} else {
			// 脱敏输出
			if len(val) > 8 {
				fmt.Printf("✅ %s = %s...%s\n", key, val[:4], val[len(val)-4:])
			} else {
				fmt.Printf("✅ %s = %s\n", key, val)
			}
		}
	}

	// 检查当前进程的 GOPATH / GOROOT
	fmt.Printf("\n=== Go 进程信息 ===\n")
	fmt.Printf("CWD: %s\n", getCWD())
	fmt.Printf("GOROOT: %s\n", os.Getenv("GOROOT"))
	fmt.Printf("GOPATH: %s\n", os.Getenv("GOPATH"))
}

func getCWD() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return cwd
}
