# 观测性学习

## 什么是观测性？

观测性就是通过系统的外部输出（日志、指标、追踪）来推断其内部状态的能力。

### 三大支柱

| 支柱 | 作用 | 典型工具 |
|------|------|---------|
| Metrics（指标） | 聚合的数值数据，用于监控性能和健康状况（如 QPS、延迟、错误率） | Prometheus, StatsD |
| Logs（日志） | 带时间戳的事件记录，用于调试和审计 | Loki, ELK |
| Traces（追踪） | 请求在分布式系统中的完整调用链路 | Jaeger, Zipkin, Tempo |

## Prometheus 的核心概念是什么？

Prometheus 是 **时序数据库 + 监控告警系统**，主要特点：

1. **Pull-based**：Prometheus 主动从目标抓取指标
2. **数据模型**：指标由 名称 + 标签 组成
3. **PromQL**：强大的查询语言，用于数据查询和分析

### 关键组件

- **Prometheus Server**：抓取、存储、查询指标
- **Exporters**：将第三方系统指标转换为 Prometheus 格式
- **Pushgateway**：用于批处理任务无法被拉取的情况
- **Alertmanager**：处理告警通知

### 数据模型示例

一个典型的 Prometheus 指标格式如下：

```promql
http_requests_total{method="POST", handler="/api/users", status="200"}
```

### 常见的指标类型

| 类型 | 用途 | 示例 |
|------|------|------|
| **Counter** | 只增不减的计数器 | 请求总数、错误数 |
| **Gauge** | 可增可减的瞬时值 | 当前 goroutine 数、内存使用量 |
| **Histogram** | 分布统计（自动计算分位数） | 请求延迟、响应大小 |
| **Summary** | 类似 Histogram，但分位数在客户端计算（较少用） | — |

> ✅ 建议：对 HTTP 服务，至少暴露：
>    - 请求总数（按方法/路径/状态码）
>    - 请求延迟（Histogram）
>    - 自定义业务指标（如订单创建数）


## Prometheus 的案例实践

### 1. 项目目录结构

```
monitor/
├── docker-compose.yaml       # Docker 编排配置
├── prometheus/
│   └── prometheus.yml        # Prometheus 配置文件
└── demo1/                    # Go 应用示例
    ├── go.mod
    ├── go.sum
    └── main.go
```

### 2. 编写 Prometheus 配置文件

在 `prometheus/prometheus.yml` 中配置：

```yaml
global:
  scrape_interval: 15s    # 抓取数据的时间间隔

scrape_configs:
  - job_name: 'demo1'
    static_configs:
      - targets: ['host.docker.internal:2365']  # Go 应用的地址
```

> **说明**：`host.docker.internal` 允许 Docker 容器访问宿主机上运行的服务。

### 3. 编写 Docker Compose 配置

```yaml
services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
    command:
      - "--config.file=/etc/prometheus/prometheus.yml"
      - "--web.enable-lifecycle" # 热重载配置
    extra_hosts:
      - "host.docker.internal:host-gateway" # 确保 Linux 兼容

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-data:/var/lib/grafana

volumes:
  grafana-data:
```

### 4. 启动服务

```bash
# 启动 Prometheus 和 Grafana
docker-compose up -d

# 启动 Go 应用（在 demo1 目录下）
cd demo1
go run main.go
```

### 5. 在 Grafana 中添加 Prometheus 数据源

1. **打开 Grafana**  
   访问 `http://localhost:3000`，使用默认用户名和密码登录（admin/admin）

2. **添加数据源**  
   - 进入 Configuration → Data Sources
   - 点击 "Add data source"
   - 选择 "Prometheus"
   - 配置 URL: `http://prometheus:9090`
   - 点击 "Save & Test"

3. **创建 Dashboard**  
   - 进入 Dashboards → New Dashboard
   - 添加面板，配置相关的查询和视图

### 6. 常用 PromQL 查询示例

```promql
# 查看请求总数
http_requests_total

# 按状态码分组查看请求总数
sum(rate(http_requests_total[5m])) by (status)

# 查看平均请求延迟
rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m])

# 查看 P95 延迟
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

---

## 访问地址

- **Prometheus UI**: http://localhost:9090
- **Grafana UI**: http://localhost:3000
- **Go 应用**: http://localhost:2365
- **指标端点**: http://localhost:2365/metrics
