# Bash 后台任务管理

> 状态: 待扩展

---

## 1. TaskManager 结构

```go
type TaskManager struct {
    mu     sync.RWMutex
    tasks  map[string]*BackgroundTask
}

type BackgroundTask struct {
    ID       string
    PID      int
    Cmd      *exec.Cmd
    State    TaskState  // Pending, Running, Failed, Exited
    LogBuf   *ring.Ring // 环形缓冲区
    Cancel   context.CancelFunc
}

type TaskState int
const (
    StatePending TaskState = iota
    StateRunning
    StateFailed
    StateExited
)
```

---

## 2. 意图判定

### 方案 A（显式声明）

在 bash 工具 schema 中增加 `isBackground bool` 字段。

### 方案 B（隐式嗅探）

检测高危/常驻命令特征：
- 关键词：`server`, `watch`, `dev`, `top`
- 后台符号：`&`, `nohup`

---

## 3. 异步拉起与脱离

```go
func (t *BashTool) executeBackground(cmdStr string) (*BackgroundTask, error) {
    cmd := exec.Command("sh", "-c", cmdStr)

    // 启动进程（不等待）
    if err := cmd.Start(); err != nil {
        return nil, err
    }

    task := &BackgroundTask{
        ID:  generateUUID(),
        PID: cmd.Process.Pid,
        Cmd: cmd,
        State: StateRunning,
        LogBuf: ring.New(1024),  // 1KB 环形缓冲
    }

    // 托管给 TaskManager
    TaskMgr.Add(task)

    // 异步接管生命周期
    go t.manageTaskLifecycle(task)

    return task, nil
}
```

---

## 4. 健康检查（2-3 秒静默期）

```go
func (t *BashTool) waitAndVerify(task *BackgroundTask) error {
    time.Sleep(3 * time.Second)

    // 检查进程是否存活
    if !isProcessRunning(task.PID) {
        return fmt.Errorf("进程启动失败: %s", task.LogBuf.String())
    }

    // 如果是网络服务，检查端口
    if isNetworkService(task) {
        if !isPortOpen(task.Port) {
            return fmt.Errorf("端口未开放: %d", task.Port)
        }
    }

    return nil
}
```

---

## 5. 环形缓冲区

使用 `container/ring` 实现固定大小日志缓冲，防止长连接撑爆内存。

---

## 6. 待实现细节

- [ ] TaskManager 的完整实现
- [ ] 意图判定（显式/隐式）
- [ ] `cmd.Start()` 异步拉起
- [ ] 2-3 秒健康检查
- [ ] 环形缓冲区日志读取