package listener

// Request
// 功能：定义请求数据结构体，包含请求类型、操作动作、目标、关联ID、时间戳、控制信息及配置信息
//
// 字段：
//
//	Type string // 请求类型，字符串类型应为
//	Action string // 操作动作，字符串类型
//	Target any // 目标，动态类型，表示请求操作的对象或数据
//	Correlation_id string // 关联ID，字符串类型，用于标识和跟踪请求unix类型
//	Timestamp string // 时间戳，字符串类型，记录请求发起的时间
//	Control string // 控制信息，字符串类型，控制指令
//	Config []any // 配置信息，动态类型数组，存储与请求相关的配置项
type Request struct {
	Type           string `json:"type"`           // 请求类型
	Action         string `json:"action"`         // 操作动作
	Target         any    `json:"target"`         // 目标
	Correlation_id string `json:"correlation_id"` // 关联ID
	Timestamp      string `json:"timestamp"`      // 时间戳
	Control        string `json:"control"`        // 控制信息
	Config         []any  `json:"config"`         // 配置信息
}

// Result
// 功能：定义测试结果结构体，包含延迟和丢包率数据
//
// 字段：
//
//	Latency []float64 // 延迟数据，浮点数数组，记录每次测试的延迟时间
//	Package_loss_rate []float64 // 丢包率数据，浮点数数组，记录每次测试的丢包率
type Result struct {
	Latency           []float64 `json:"latency"`           // 延迟数据
	Package_loss_rate []float64 `json:"package_loss_rate"` // 丢包率数据
}

// Error
// 功能：定义错误信息结构体，包含错误代码和详细描述
//
// 字段：
//
//	Code int // 错误代码，整数类型
//	Message string // 错误消息，字符串类型，提供错误的详细描述
type Error struct {
	Code    int    `json:"code"`    // 错误代码
	Message string `json:"message"` // 错误消息
}

// Response
// 功能：定义响应数据结构体，包含请求类型、操作动作、配置信息、目标列表、状态、结果、关联ID、时间戳、控制信息及可能存在的错误详情
//
// 字段：
//
//	Type string // 请求类型，字符串类型
//	Action string // 操作动作，字符串类型
//	Config []any // 配置信息，动态类型数组，存储与请求相关的配置项
//	Target []string // 目标列表，字符串数组，记录请求涉及的目标对象
//	Status string // 响应状态，字符串类型，表示请求处理的状态
//	Result map[string]any // 处理结果，键值对形式的动态类型映射，存储请求处理的具体结果数据
//	Correlation_id string // 关联ID，字符串类型，用于标识和跟踪请求
//	Timestamp string // 时间戳，字符串类型，记录响应生成的时间
//	Control string // 控制信息，字符串类型，可能包含额外的控制指令或状态信息
//	Error Error // 错误信息（如果存在），Error结构体类型，详细描述发生的问题
type Response struct {
	Type           string         `json:"type"`           // 请求类型
	Action         string         `json:"action"`         // 操作动作
	Config         []any          `json:"config"`         // 配置信息
	Target         []string       `json:"target"`         // 目标列表
	Status         string         `json:"status"`         // 响应状态
	Result         map[string]any `json:"result"`         // 处理结果
	Correlation_id string         `json:"correlation_id"` // 关联ID
	Timestamp      string         `json:"timestamp"`      // 时间戳
	Control        string         `json:"control"`        // 控制信息
	Error          Error          `json:"error"`          // 错误信息（如果存在）
}

// RunResult
// 功能：定义运行结果结构体，包含类型、关联ID、时间戳及状态信息
//
// 字段：
//
//	Type string // 结果类型，字符串类型
//	Corrlation_id string // 关联ID，字符串类型，用于标识和跟踪请求
//	Timestamp string // 时间戳，字符串类型，记录运行结果生成的时间
//	Status string // 运行状态，字符串类型，表示运行任务的最终状态
type RunResult struct {
	Type          string `json:"type"`           // 结果类型
	Corrlation_id string `json:"correlation_id"` // 关联ID
	Timestamp     string `json:"timestamp"`      // 时间戳
	Status        string `json:"status"`         // 运行状态
}
