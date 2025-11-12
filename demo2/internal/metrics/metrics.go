package metrics

import "github.com/prometheus/client_golang/prometheus"

// Prometheus 指标类型说明：
// 1. Counter（计数器）：只能递增的累加器，用于统计事件发生的总次数
//    - 特点：只增不减，重启后归零
//    - 适用场景：请求总数、错误总数、任务完成数等
//    - 查询：通常使用 rate() 或 increase() 函数计算速率
//
// 2. Gauge（仪表盘）：可以任意增减的数值
//    - 特点：可增可减，反映当前状态
//    - 适用场景：当前在线用户数、内存使用量、队列长度等
//    - 查询：直接使用当前值，或使用 avg/min/max 等聚合函数
//
// 3. Histogram（直方图）：对观测值进行采样，统计分布情况
//    - 特点：自动统计观测值的分布区间（buckets）、总和（sum）、计数（count）
//    - 适用场景：请求耗时、响应大小等需要了解分布的指标
//    - 查询：可计算百分位数 histogram_quantile()
//
// 4. Summary（摘要）：类似 Histogram，但在客户端计算分位数
//    - 特点：直接提供分位数，但无法聚合多个实例的数据
//    - 适用场景：与 Histogram 类似，但更关注精确分位数

// 命名空间建议：{service_name}_{metric_name}
const namespace = "user_service"

// HTTP 指标
var (
	// HTTPRequestsTotal 使用 Counter 类型
	// 记录 HTTP 请求的累计总数，按方法、端点、状态码分组
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTPRequestDuration 使用 Histogram 类型
	// 记录 HTTP 请求的耗时分布，可用于计算 P50、P95、P99 等百分位数
	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "Duration of HTTP requests in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

// Bussiness 指标
var (
	// UserRegistrationsTotal 使用 Counter 类型
	// 记录用户注册的累计总数，按状态（成功/失败）分组
	UserRegistrationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "user_registrations_total",
			Help:      "Total number of user registrations",
		},
		[]string{"status"},
	)

	// UserRegistrationDuration 使用 Histogram 类型
	// 记录用户注册操作的耗时分布，用于监控注册性能
	UserRegistrationDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "user_registration_duration_seconds",
			Help:      "Duration of user registrations in seconds",
			Buckets:   prometheus.DefBuckets,
		},
	)

	// ActiveUsers 使用 Gauge 类型
	// 记录当前活跃用户数，可随时增加或减少，反映实时状态
	ActiveUsers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "active_users",
			Help:      "Current number of active users",
		},
	)
)
