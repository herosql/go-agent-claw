# 工具注册与分发机制

> 状态: 待扩展

---

## 1. 路径穿越防护

### 方案一：纯代码层路径安全锚定（Chroot 语义模拟）

```go
func isPathSafe(workDir, requestedPath string) bool {
    // 解析最终绝对路径（处理软链接）
    absPath, err := filepath.EvalSymlinks(requestedPath)
    if err != nil {
        return false
    }

    // 检查是否在 workDir 以内
    absWorkDir, _ := filepath.EvalSymlinks(workDir)
    return strings.HasPrefix(absPath, absWorkDir)
}
```

### 方案二：OS 级沙箱

- **chroot / pivot_root**: 将进程根目录更改为 workDir
- **低权限用户**: 用 `agent_runner` 用户启动 Agent 进程
- **ACL 权限控制**: 仅授予特定目录读写权限

### 方案三：容器化（工业级）

- 动态拉起 Docker/轻量级虚拟机沙箱
- 只挂载需要的项目代码
- 任务结束后物理销毁容器

---

## 2. 工具输出卸载（Tool Call Offloading）

当工具输出超过阈值时：
1. 将完整内容写入磁盘临时目录
2. 返回摘要消息：`"文件过长（共 5000 行，已卸载至 /tmp/xxx）"`
3. 全局 Compactor 仍监控 Token 水位

---

## 3. 待实现细节

- [ ] `isPathSafe` 函数实现与测试
- [ ] OS 级沙箱支持（chroot/低权限用户）
- [ ] 容器化沙箱支持
- [ ] 工具输出卸载机制