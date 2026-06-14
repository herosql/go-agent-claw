// internal/observability/tracker.go
package observability

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	ctxpkg "github.com/herosql/go-agent-claw/internal/context"
	"github.com/herosql/go-agent-claw/internal/provider"
	"github.com/herosql/go-agent-claw/internal/schema"
)

func formatInt(n int) string      { return strconv.Itoa(n) }
func formatCost(f float64) string { return strconv.FormatFloat(f, 'f', 6, 64) }

// PricingModel 定义了不同大模型的计费标准 (单位: 美元/1M Tokens)
// 为了演示，这里硬编码了当前市面上几个主流模型的官方大致定价。
var PricingModel = map[string]struct {
	InputPrice  float64
	OutputPrice float64
}{
	"glm-4.5-air": {InputPrice: 0.15, OutputPrice: 0.15}, // 这里假定的大模型价格(每百万Token，tk)
}

// CostTracker 是一个包装了真实 LLMProvider 的装饰器中间件
type CostTracker struct {
	nextProvider provider.LLMProvider
	session      *ctxpkg.Session
	modelName    string
}

// NewCostTracker 构造函数：接收一个现有的 Provider，返回一个被监控的 Provider
func NewCostTracker(next provider.LLMProvider, modelName string, session *ctxpkg.Session) *CostTracker {
	return &CostTracker{
		nextProvider: next,
		modelName:    modelName,
		session:      session,
	}
}

// Generate 实现了 LLMProvider 接口！这意味着它可以被无缝注入到 Main Loop 中。
func (t *CostTracker) Generate(ctx context.Context, msgs []schema.Message, availableTools []schema.ToolDefinition) (*schema.Message, error) {

	// 1. 记录请求发起的时刻
	startTime := time.Now()

	// 2. 调用真实的底层大模型去执行耗时的网络请求
	respMsg, err := t.nextProvider.Generate(ctx, msgs, availableTools)

	// 3. 计算耗时
	latency := time.Since(startTime)

	// 如果报错了，只打印报错时间，不计费
	if err != nil {
		slog.Info("[Tracker] ❌ API 调用失败 | 耗时: " + latency.String())
		return respMsg, err
	}

	// 4. 解析 Token 并计算成本
	if respMsg.Usage != nil {
		promptTokens := respMsg.Usage.PromptTokens
		completionTokens := respMsg.Usage.CompletionTokens

		var cost float64
		if price, exists := PricingModel[t.modelName]; exists {
			// 计算美元花费 = (输入Tokens * 输入单价 + 输出Tokens * 输出单价) / 1000000
			cost = (float64(promptTokens)*price.InputPrice + float64(completionTokens)*price.OutputPrice) / 1000000.0
		}

		// 5. 打印精美的仪表盘日志
		slog.Info("[Tracker] 📊 API 调用完成 | 耗时: " + latency.String() +
			" | 输入: " + formatInt(promptTokens) + " tk | 输出: " + formatInt(completionTokens) +
			" tk | 花费: ¥" + formatCost(cost))

		// 6. 将账单累加到当前的 Session 中，供人类后续随时查询
		if t.session != nil {
			t.session.RecordUsage(promptTokens, completionTokens, cost)
			slog.Info("[Tracker] 💰 当前会话 (" + t.session.ID + ") 累计花费: ¥" + formatCost(t.session.TotalCostCNY))
		}
	} else {
		slog.Info("[Tracker] ⚠️ API 调用完成，但未返回 Usage 数据 | 耗时: " + latency.String())
	}

	return respMsg, nil
}
