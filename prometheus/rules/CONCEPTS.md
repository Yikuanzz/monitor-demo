# Prometheus 核心概念详解

## 📊 一、Recording Rules（记录规则）

### 什么是 Recording Rules？

预先计算并存储复杂的 PromQL 查询结果，生成新的时间序列指标。

### 为什么需要？

```shell
原始查询（慢）:
  histogram_quantile(0.99, sum by(le) (rate(http_duration_bucket[5m])))
  ↓ 每次查询都要实时计算

Recording Rule（快）:
  http_request_duration_seconds_p99  ← 提前算好，直接读取
```

### 典型使用场景

- ✅ **性能优化**：Dashboard 频繁查询的复杂指标
- ✅ **告警规则**：避免告警评估时计算超时
- ✅ **数据聚合**：跨多个服务聚合统计
- ✅ **简化查询**：复杂逻辑封装成简单指标名

---

## 📈 二、PromQL 常用函数

### 1. `rate()` - 计算增长率

```promql
rate(user_service_http_requests_total[5m])
```

- **作用**：计算 Counter 在时间窗口内的每秒平均增长率
- **输入**：Counter 类型指标
- **输出**：每秒增量（float）
- **时间窗口**：`[5m]` 表示最近 5 分钟
- **使用场景**：计算 QPS、错误率、吞吐量

**示例**：

```shell
假设 5 分钟内请求从 1000 增加到 2000
rate = (2000 - 1000) / 300秒 = 3.33 req/s
```

### 2. `sum by()` - 按标签分组聚合

```promql
sum by (method, endpoint) (
  rate(http_requests_total[5m])
)
```

- **作用**：按指定标签分组求和，保留这些标签维度
- **保留标签**：`method`, `endpoint`
- **丢弃标签**：其他所有标签（如 `instance`, `job`）
- **使用场景**：跨实例聚合、按维度统计

**对比**：

```promql
sum(...)              # 全部求和，不保留任何标签
sum by (endpoint)(...)  # 按 endpoint 分组，只保留 endpoint 标签
```

### 3. `histogram_quantile()` - 计算百分位数

```promql
histogram_quantile(0.99,
  sum by (le) (
    rate(http_duration_seconds_bucket[5m])
  )
)
```

- **作用**：基于 Histogram 的 bucket 数据计算分位数
- **参数**：`0.99` = P99，`0.95` = P95，`0.50` = P50（中位数）
- **必需标签**：`le`（less than or equal，Histogram 自动生成）
- **使用场景**：延迟分析、性能监控

---

## 🎯 三、P99 延迟深度解析

### 什么是 P99？

**P99（99th Percentile，第 99 百分位数）**：表示 99% 的请求响应时间都 ≤ 这个值。

### 直观理解

```shell
假设有 100 个请求的响应时间（按从小到大排序）：
[10ms, 15ms, 20ms, ..., 450ms, 500ms]
              ↑
         P99 ≈ 450ms

含义：99 个请求 ≤ 450ms，只有 1 个请求 > 450ms
```

### 为什么关注 P99 而非平均值？

| 指标 | 优点 | 缺点 | 适用场景 |
|------|------|------|----------|
| **平均值** | 计算简单 | 被极端值严重影响 | 初步评估 |
| **P50（中位数）** | 反映大多数情况 | 忽略一半用户体验 | 一般性能 |
| **P99** | 发现长尾问题 | 对异常敏感 | **生产监控** |
| **P999** | 极端情况监控 | 样本要求高 | 高可用系统 |

### 真实案例对比

```shell
场景：100 个请求
- 99 个请求：50ms
- 1 个请求：5000ms（超时）

平均值 = (99*50 + 5000) / 100 = 99.5ms  ← 看起来很好！
P99 = 5000ms                           ← 实际有严重问题！
```

### P99 计算原理

Histogram 将观测值分桶存储：

```shell
Bucket(le=0.1):  50 个   ← 50 个请求 ≤ 0.1s
Bucket(le=0.5):  95 个   ← 95 个请求 ≤ 0.5s
Bucket(le=1.0):  99 个   ← 99 个请求 ≤ 1.0s   ← 这就是 P99！
Bucket(le=5.0):  100 个  ← 100 个请求 ≤ 5.0s
```

---

## 🔍 四、Histogram 指标详解

### Histogram 自动生成的时间序列

当你定义一个 Histogram：

```go
HTTPRequestDuration = prometheus.NewHistogram(...)
```

Prometheus 会自动生成 3 个时间序列：

#### 1. `_bucket` - 分桶计数（用于 P99 计算）

```shell
http_request_duration_seconds_bucket{le="0.1"} = 50
http_request_duration_seconds_bucket{le="0.5"} = 95
http_request_duration_seconds_bucket{le="1.0"} = 99
http_request_duration_seconds_bucket{le="+Inf"} = 100  ← 总数
```

#### 2. `_sum` - 总和（用于计算平均值）

```shell
http_request_duration_seconds_sum = 120.5  ← 所有请求耗时总和
```

#### 3. `_count` - 总数

```shell
http_request_duration_seconds_count = 100  ← 请求总数
```

### 计算平均值

```promql
rate(http_request_duration_seconds_sum[5m])
  /
rate(http_request_duration_seconds_count[5m])
```

---

## 🎨 五、标签选择器语法

### 基本选择器

```promql
# 精确匹配
{status="200"}

# 不等于
{status!="200"}

# 正则匹配
{status=~"2..|3.."}     # 匹配 2xx 或 3xx
{status=~"5..|4.."}     # 匹配 4xx 或 5xx（错误）

# 正则不匹配
{endpoint!~"/health.*"} # 排除健康检查端点

# 非空
{endpoint!=""}          # 排除空值
```

### 正则语法说明

```shell
.   → 匹配任意单个字符
.*  → 匹配任意多个字符
5.. → 匹配 5 开头的三位数（如 500, 502, 503）
^   → 开头
$   → 结尾
|   → 或
```

---

## 📐 六、时间窗口选择

### 窗口大小建议

```promql
rate(metric[1m])   # 过短：数据噪音大，波动剧烈
rate(metric[5m])   # 推荐：平衡响应速度和稳定性
rate(metric[15m])  # 较长：更平滑，但反应慢
```

### 经验法则

- **告警规则**：`[5m]` 或 `[10m]`（避免误报）
- **实时监控**：`[1m]` 或 `[2m]`（快速发现问题）
- **趋势分析**：`[1h]` 或 `[6h]`（长期趋势）

---

## 🚀 七、实战示例

### 示例 1：计算 API 错误率

```promql
# 错误请求数 / 总请求数
sum(rate(http_requests_total{status=~"5.."}[5m]))
  /
sum(rate(http_requests_total[5m]))
```

### 示例 2：计算 P95、P99 对比

```promql
# P95
histogram_quantile(0.95,
  sum by (le) (rate(http_duration_bucket[5m]))
)

# P99
histogram_quantile(0.99,
  sum by (le) (rate(http_duration_bucket[5m]))
)
```

### 示例 3：按端点统计 TOP 5 慢查询

```promql
topk(5,
  histogram_quantile(0.99,
    sum by (endpoint, le) (
      rate(http_duration_bucket[5m])
    )
  )
)
```

### 示例 4：计算服务可用性（SLA）

```promql
# 成功率（非 5xx）
sum(rate(http_requests_total{status!~"5.."}[5m]))
  /
sum(rate(http_requests_total[5m]))
  * 100
```

---

## 📚 八、命名规范

### Recording Rule 命名

推荐格式：`level:metric:operations`

```yaml
# ✅ 好的命名
- record: job:http_requests:rate5m              # 任务级别
- record: instance:cpu_usage:avg               # 实例级别
- record: http_error_ratio:5m                  # 错误率
- record: http_request_duration_seconds_p99    # P99 延迟

# ❌ 不好的命名
- record: my_metric                           # 太模糊
- record: http_requests_rate_5min             # 不符合规范
```

---

## 🎓 九、常见问题 FAQ

### Q1: rate() 和 increase() 有什么区别？

```promql
rate(metric[5m])      # 返回每秒速率（float）
increase(metric[5m])  # 返回总增量（= rate * 300）
```

### Q2: 为什么 histogram_quantile 需要 rate()？

因为 `_bucket` 是累计计数（Counter），需要先计算增长率才能反映时间窗口内的分布。

### Q3: P99 = 0 是什么意思？

说明 99% 的请求响应时间都在最小的 bucket 范围内，性能极好。

### Q4: 如何选择 Histogram 的 buckets？

```go
// 默认 buckets（适合秒级延迟）
prometheus.DefBuckets  // [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]

// 毫秒级延迟
prometheus.LinearBuckets(0, 10, 10)  // [0, 10, 20, ..., 90] ms

// 自定义
Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 5}
```

---

## 🛠️ 十、调试技巧

### 1. 查看 Histogram 的 bucket 分布

```promql
http_request_duration_seconds_bucket
```

### 2. 验证 Recording Rule 是否生效

访问：`http://localhost:9090/rules`

### 3. 测试 PromQL 表达式

使用 Prometheus Web UI 的 Graph 页面

### 4. 查看指标的所有标签
```promql
{__name__=~"http_.*"}
```

---

## 📖 相关资源

- [Prometheus 官方文档](https://prometheus.io/docs/)
- [PromQL 查询语言](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Histogram vs Summary](https://prometheus.io/docs/practices/histograms/)
- [Recording Rules 最佳实践](https://prometheus.io/docs/practices/rules/)

---

**提示**：这些概念环环相扣，建议结合实际代码和 Prometheus UI 进行实践！
