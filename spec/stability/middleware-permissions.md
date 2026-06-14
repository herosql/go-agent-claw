# 防御纵深：利用 Middleware 实现高危命令拦截与飞书人工审批

> 状态: 待扩展

---

## 背景与问题

"代码硬编码"的高危命令拦截不及格。如果明天运维要求把 `kubectl delete` 加入拦截名单，不能去修改 Go 源码、重新编译、重启引擎。需要支持外部配置 + 运行时热更新。

---

## 核心方案：三态权限与动态热加载

### 1. 模型设计：三态权限与组织架构矩阵（Schema）

**配置文件**：`~/.claw/permissions.yaml`

```yaml
version: "1.0"

globaldenypatterns:  # 全局无条件封杀（Deny）
  - rm\s+-rf\s+/
  - drop\s+database

permission_rules:
  - id: "rule_001"
    role: "junior_dev"       # 研发初级角色
    tools: ["bash"]
    scope: ["prod_server"]    # 涉及生产环境范围
    action_match: .*          # 只要是物理命令
    policy: "ask"             # 必须触发人工审批
    approvers: ["tech_lead"]  # 审批人指定为技术主管

  - id: "rule_002"
    role: "senior_dev"        # 高级研发
    tools: ["bash"]
    scope: ["stg_server"]     # 测试环境范围
    action_match: "^go\\s+test|^git\\s+"
    policy: "allow"           # 允许直接 YOLO 放行

  - id: "rule_003"
    role: "intern"            # 实习生
    tools: ["writefile", "editfile"]
    scope: ["*"]
    action_match: .*
    policy: "deny"            # 一律无情拒绝
```

### 2. 动态加载层：高并发下的"读写锁与双缓冲"热加载

**并发控制机制（Go 语义层设计）**：

```go
type PermissionEngine struct {
    mu      sync.RWMutex
    rules   []PermissionRule
    engine  *RuleEngine  // 双缓冲：切换时原子替换
}

func (p *PermissionEngine) Check(ctx context.Context, tool, role, scope string) Policy {
    p.mu.RLock()
    defer p.mu.RUnlock()
    return p.engine.Eval(tool, role, scope)
}
```

**热加载触发**：
- 文件系统：`fsnotify` 监听文件变更事件
- API 控制：提供 `POST /api/v1/permissions/reload` 接口
- 防抖动（Debounce）：避免编辑器连续 Ctrl+S 导致频繁写锁重载

**双缓冲切换**：
```go
func (p *PermissionEngine) Reload(newRules []PermissionRule) {
    p.mu.Lock()
    defer p.mu.Unlock()
    // 内存中完整构建新引擎，纳秒级指针赋值切换
    p.engine = NewRuleEngine(newRules)
}
```

### 3. 分流层：Middleware 层的动态路由分发

**执行链流转逻辑**：

```
工具调用 → Middleware 拦截
  → 身份提取（从 Context 提取 UserId、Role）
  → 全局黑名单碰撞（命中直接 deny）
  → 矩阵规则匹配
      ├── allow → 释放读锁，沉降到工具 Execute
      ├── deny → 返回个性化拒绝理由
      └── ask → 挂起协程，发送飞书卡片给 approvers
        → 收到 approve {TaskID} → 唤醒协程放行
```

---

## 核心数据结构

```go
type PermissionRule struct {
    ID           string   `yaml:"id"`
    Role         string   `yaml:"role"`
    Tools        []string `yaml:"tools"`
    Scope        []string `yaml:"scope"`
    ActionMatch  string   `yaml:"action_match"`
    Policy       string   `yaml:"policy"`  // allow / ask / deny
    Approvers    []string `yaml:"approvers,omitempty"`
}

type Policy string

const (
    PolicyAllow Policy = "allow"
    PolicyAsk   Policy = "ask"
    PolicyDeny  Policy = "deny"
)
```

---

## 待扩展

### 待扩展

> 待补充详细设计方案

1. **飞书卡片渲染**：审批卡片的模板设计与交互细节
2. **超时机制**：人工审批超时后的降级策略

### 待扩展2

> 未来规划方向

1. **规则版本管理**：支持规则的灰度发布与回滚
2. **审计日志**：所有拦截与审批行为的完整记录