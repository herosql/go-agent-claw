// internal/tools/edit_file.go
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/herosql/go-agent-claw/internal/schema"
)

type EditFileTool struct {
	workDir string
}

func NewEditFileTool(workDir string) *EditFileTool {
	return &EditFileTool{workDir: workDir}
}

func (t *EditFileTool) Name() string {
	return "edit_file"
}

func (t *EditFileTool) Definition() schema.ToolDefinition {
	return schema.ToolDefinition{
		Name:        t.Name(),
		Description: "对现有文件进行局部的字符串替换。这比重写整个文件更安全、更快速。请提供足够的 old_text 上下文以确保匹配的唯一性。",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "要修改的文件路径",
				},
				"old_text": map[string]interface{}{
					"type":        "string",
					"description": "文件中原有的文本。必须包含足够的上下文（建议上下各多包含几行），以确保在文件中的唯一性。",
				},
				"new_text": map[string]interface{}{
					"type":        "string",
					"description": "要替换成的新文本",
				},
			},
			"required": []string{"path", "old_text", "new_text"},
		},
	}
}

type editFileArgs struct {
	Path    string `json:"path"`
	OldText string `json:"old_text"`
	NewText string `json:"new_text"`
}

// internal/tools/edit_file.go (续)

// isPathSafe 检查路径是否在允许的工作区内，防止路径穿越攻击
func isPathSafe(fullPath, workDir string) error {
	// 禁止空字节
	if strings.Contains(fullPath, "\x00") {
		return fmt.Errorf("路径包含非法字符，不允许操作")
	}

	// 清理并解析路径
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("工作区路径无效: %w", err)
	}
	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return fmt.Errorf("路径解析失败: %w", err)
	}

	// 检查解析后的路径是否在工作区目录内
	if !strings.HasPrefix(absFullPath, absWorkDir) {
		return fmt.Errorf("路径越界，不允许访问工作区外的文件")
	}

	return nil
}

// isUserPathSafe 检查用户提供的原始路径，优先拦截绝对路径和穿越路径
func isUserPathSafe(userPath string) error {
	// 禁止空字节
	if strings.Contains(userPath, "\x00") {
		return fmt.Errorf("路径包含非法字符，不允许操作")
	}

	// 禁止绝对路径
	if filepath.IsAbs(userPath) {
		return fmt.Errorf("不允许使用绝对路径，仅支持相对路径")
	}

	// 禁止包含 .. 路径穿越
	cleaned := filepath.Clean(userPath)
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("路径不允许包含 .. 穿越上级目录")
	}

	return nil
}

// fuzzyReplace 实现了四级容错降级替换算法
func fuzzyReplace(originalContent, oldText, newText string) (string, error) {
	// L1: 精确匹配，使用 Index 定位而非 Count（避免重叠计数误判）
	idx := strings.Index(originalContent, oldText)
	if idx >= 0 {
		// 检查是否唯一：查找 idx 之后是否还有匹配
		if strings.Contains(originalContent[idx+len(oldText):], oldText) {
			return "", fmt.Errorf("old_text 匹配到了多处，请提供更多的上下文代码以确保唯一性")
		}
		return replaceAt(originalContent, idx, idx+len(oldText), newText), nil
	}

	// L2: 换行符归一化 (统一将 \r\n 转换为 \n)
	normalizedContent := strings.ReplaceAll(originalContent, "\r\n", "\n")
	normalizedOld := strings.ReplaceAll(oldText, "\r\n", "\n")

	idx = strings.Index(normalizedContent, normalizedOld)
	if idx >= 0 {
		if strings.Contains(normalizedContent[idx+len(normalizedOld):], normalizedOld) {
			return "", fmt.Errorf("old_text 匹配到了多处，请提供更多的上下文代码以确保唯一性")
		}
		return replaceAt(originalContent, idx, idx+len(normalizedOld), newText), nil
	}

	// L3: Trim Space 匹配 (忽略首尾的空行和空格)
	trimmedOld := strings.TrimSpace(normalizedOld)
	if trimmedOld != "" {
		idx = strings.Index(normalizedContent, trimmedOld)
		if idx >= 0 {
			if strings.Contains(normalizedContent[idx+len(trimmedOld):], trimmedOld) {
				return "", fmt.Errorf("old_text 匹配到了多处，请提供更多的上下文代码以确保唯一性")
			}
			// L3 替换时，必须在原始 normalizedContent 中找到实际位置进行替换，
			// 以保留该位置前后的所有内容（包括行首缩进和行尾字符）
			return replaceAt(originalContent, idx, idx+len(trimmedOld), newText), nil
		}
	}

	// L4: 逐行去缩进匹配 (最强力的容错：消除大模型遗漏缩进的幻觉)
	return lineByLineReplace(originalContent, oldText, newText)
}

// replaceAt 用 newText 替换 original 中 [start:end) 的内容，
// 注意：start/end 是基于已归一化文本的位置，original 须保证长度一致
func replaceAt(original string, start, end int, newText string) string {
	return original[:start] + newText + original[end:]
}

// lineByLineReplace 将文本按行切割，去除首尾空白后进行滑动窗口匹配
func lineByLineReplace(content, oldText, newText string) (string, error) {
	// 统一归一化换行符，与 L2 保持一致，确保 \r\n 文件也能正确处理
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	normalizedOld := strings.ReplaceAll(oldText, "\r\n", "\n")

	contentLines := strings.Split(normalized, "\n")
	oldLines := strings.Split(strings.TrimSpace(normalizedOld), "\n")

	if len(oldLines) == 0 || len(contentLines) < len(oldLines) {
		return "", fmt.Errorf("找不到该代码片段")
	}

	// 清理 oldLines 的每行首尾空白
	for i := range oldLines {
		oldLines[i] = strings.TrimSpace(oldLines[i])
	}

	matchCount := 0
	matchStartIndex := -1
	matchEndIndex := -1

	// 滑动窗口在原始文件中寻找匹配块
	for i := 0; i <= len(contentLines)-len(oldLines); i++ {
		isMatch := true
		for j := 0; j < len(oldLines); j++ {
			if strings.TrimSpace(contentLines[i+j]) != oldLines[j] {
				isMatch = false
				break
			}
		}

		if isMatch {
			matchCount++
			matchStartIndex = i
			matchEndIndex = i + len(oldLines)
		}
	}

	if matchCount == 0 {
		return "", fmt.Errorf("在文件中未找到 old_text，请大模型先调用 read_file 仔细确认文件内容和缩进")
	}
	if matchCount > 1 {
		return "", fmt.Errorf("模糊匹配到了 %d 处相似代码，请提供更多上下行代码以精确定位", matchCount)
	}

	// 将 newText 按 \n 拆分为多行后再插入
	newTextLines := strings.Split(newText, "\n")

	// 构建新内容：matchStartIndex 之前的行 + newText 拆分的行 + matchEndIndex 之后的行
	var newContentLines []string
	newContentLines = append(newContentLines, contentLines[:matchStartIndex]...)
	newContentLines = append(newContentLines, newTextLines...)
	newContentLines = append(newContentLines, contentLines[matchEndIndex:]...)

	return strings.Join(newContentLines, "\n"), nil
}

// internal/tools/edit_file.go (续)

func (t *EditFileTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var input editFileArgs
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	// 【安全防线】检查用户提供的路径是否为绝对路径或包含路径穿越
	if err := isUserPathSafe(input.Path); err != nil {
		return "", err
	}

	fullPath := filepath.Join(t.workDir, input.Path)

	// 【安全防线】路径安全检查：防止路径穿越攻击
	if err := isPathSafe(fullPath, t.workDir); err != nil {
		return "", err
	}

	// 1. 读取原文件内容
	contentBytes, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败，请确认路径是否正确: %w", err)
	}
	originalContent := string(contentBytes)

	// 2. 调用多级模糊替换算法
	newContent, err := fuzzyReplace(originalContent, input.OldText, input.NewText)
	if err != nil {
		// 【驾驭哲学】将具体的报错原因 (如匹配到多处) 原样返回，让大模型自行纠正
		return "", err
	}

	// 3. 将新内容安全地写回磁盘
	if err := os.WriteFile(fullPath, []byte(newContent), 0600); err != nil {
		return "", fmt.Errorf("写回文件失败: %w", err)
	}

	return fmt.Sprintf("✅ 成功修改文件: %s", input.Path), nil
}
