# 环境检查清单

## 1. 运行时依赖

### Go 工具链
| 检查项 | 要求版本 | 验证命令 |
|---|---|---|
| Go 编译器 | ≥ 1.26 | `go version` |
| GOPATH/bin 在 PATH | — | `go env GOPATH` |
| 模块代理（国内） | 可选 | `go env GOPROXY`（推荐 `https://goproxy.cn,direct`）|

### 外部 API Key（必须）
| 环境变量 | 说明 | 验证命令 |
|---|---|---|
| `ZHIPU_API_KEY` | 智谱 GLM 模型 Key | `echo $ZHIPU_API_KEY`（非空）|

### 飞书集成（如需飞书 Bot）
| 环境变量 | 说明 | 验证命令 |
|---|---|---|
| `FEISHU_APP_ID` | 飞书应用 App ID | `echo $FEISHU_APP_ID`（非空）|
| `FEISHU_APP_SECRET` | 飞书应用 App Secret | `echo $FEISHU_APP_SECRET`（非空）|
| `FEISHU_VERIFY_TOKEN` | HTTP Webhook 验签 Token（仅 cmd/feishu）| 可选 |
| `FEISHU_ENCRYPT_KEY` | 消息加密密钥（仅 cmd/feishu）| 可选 |

---

## 2. Go 依赖安装

```bash
# 进入项目根目录
cd e:/project/go-project/go-agent-claw

# 下载并缓存所有依赖（首次或每次 go.mod 变更后）
go mod download

# 验证所有依赖可解析
go mod tidy
```

---

## 3. 编译验证

```bash
# 编译所有可执行文件（无报错即通过）
go build ./...

# 各入口独立验证
go build -o bin/claw   ./cmd/claw
go build -o bin/feishu ./cmd/feishu
go build -o bin/agentops ./cmd/agentops
go build -o bin/bench   ./cmd/bench
go build -o bin/envcheck ./cmd/envcheck
```

---

## 4. 环境诊断工具

```bash
# 快速诊断所有环境变量
go run ./cmd/envcheck
```

期望输出：所有 `✅` 状态，无 `❌`。

---

## 5. 可选：飞书 Bot 本地开发环境

### 飞书开发者控制台配置
1. 登录 [飞书开放平台](https://open.feishu.cn)
2. 创建企业自建应用，获取 `App ID` 和 `App Secret`
3. 配置权限：`im:message`、`im:message:receive`
4. 配置事件订阅：
   - HTTP 模式：请求 URL 为 `http://<your-host>:48080/webhook/event`
   - WebSocket 模式：直接使用 lark-sdk 的 WebSocket 长连接
5. 发布应用（开发阶段可使用测试版本）

### 本地网络暴露（如使用 HTTP Webhook）
飞书平台要求 HTTP Webhook 可公网访问，开发阶段可使用 [ngrok](https://ngrok.com)：
```bash
ngrok http 48080
# 将输出的 https URL 配置到飞书开发者控制台的事件订阅 URL
```

---

## 6. 依赖服务（无自建服务）

本项目无数据库或消息队列等自建依赖服务，所有外部调用均通过 HTTP：
- 智谱 GLM API（需网络可达，默认指向 `https://open.bigmodel.cn`）
- 飞书开放平台 API（需网络可达）

---

## 7. 本地开发工具（可选）

| 工具 | 用途 |
|---|---|
| [ngrok](https://ngrok.com) | HTTP Webhook 本地调试 |
| [TaskExplorer](https://github.com/go-task/task) | `task` 替代 Makefile（已在 Makefile 缺失时可用）|
| [Air](https://github.com/cosmtrek/air) | Go 热重载（`air run ./cmd/claw`）|

---

## 8. 快速验证清单（copy-paste）

```bash
# 1行命令验证全部
cd e:/project/go-project/go-agent-claw && \
  go version && \
  go mod download && \
  go build ./... && \
  go run ./cmd/envcheck
```

全部 `✅` 即环境就绪。