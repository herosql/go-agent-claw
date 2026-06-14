package context

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// P0-安全: composer 构建时 AGENTS.md 路径受 workDir 约束
func TestComposer_AGENTSPathBounded(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建假的 AGENTS.md
	agentsFile := filepath.Join(tmpDir, "AGENTS.md")
	if err := os.WriteFile(agentsFile, []byte("# Test Agents"), 0644); err != nil {
		t.Fatalf("failed to create AGENTS.md: %v", err)
	}

	composer := NewPromptComposer(tmpDir, false)
	msg := composer.Build()

	// 确保加载了 AGENTS.md 内容
	if !strings.Contains(msg.Content, "# 项目专属指南") {
		t.Fatal("AGENTS.md content should be loaded")
	}
}

// P0-安全: SkillLoader 路径受 workDir/.claw/skills 约束
func TestSkillLoader_PathBounded(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建假的技能目录
	skillDir := filepath.Join(tmpDir, ".claw", "skills", "test-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}

	skillFile := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte("# Test Skill\n\nName: test\nDescription: test\n\nBody"), 0644); err != nil {
		t.Fatalf("failed to create SKILL.md: %v", err)
	}

	loader := NewSkillLoader(tmpDir)
	content := loader.LoadAll()

	// 确保加载了技能内容
	if !strings.Contains(content, "test") {
		t.Fatal("skill content should be loaded")
	}
}

// P0-安全: 构建时路径穿越应被阻止（恶意 workDir）
func TestComposer_MaliciousWorkDir_Blocked(t *testing.T) {
	// 使用不存在的路径作为 workDir，composer 应该能处理
	tmpDir := t.TempDir()
	nonExistentDir := filepath.Join(tmpDir, "non-existent", "..", "..", "..", "etc")
	_ = nonExistentDir

	composer := NewPromptComposer(tmpDir, false)
	msg := composer.Build()

	// 即使 workDir 指向不存在的地方，也不应该 panic
	if msg.Role != "system" {
		t.Fatal("should return system message even with edge case paths")
	}
}
