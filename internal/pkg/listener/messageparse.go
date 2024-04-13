package listener

type Request struct {
	Type           string `json:"type"`
	Action         string `json:"action"`
	Target         any    `json:"target"`
	Correlation_id string `json:"correlation_id"`
	Timestamp      string `json:"timestamp"`
	Control        string `json:"control"`
	Config         []any  `json:"config"`
}
type Result struct {
	Latency           []float64 `json:"latency"`
	Package_loss_rate []float64 `json:"package_loss_rate"`
}
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type Response struct {
	Type           string         `json:"type"`
	Action         string         `json:"action"`
	Config         []any          `json:"config"`
	Target         []string       `json:"target"`
	Status         string         `json:"status"`
	Result         map[string]any `json:"result"`
	Correlation_id string         `json:"correlation_id"`
	Timestamp      string         `json:"timestamp"`
	Control        string         `json:"control"`
	Error          Error          `json:"error"`
}
type RunResult struct {
	Type string `json:"type"`

	Corrlation_id string `json:"correlation_id"`
	Timestamp     string `json:"timestamp"`
	Status        string `json:"status"`
}
