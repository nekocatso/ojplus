package listener

type Request struct {
	Type           string `json:"type"`
	Action         string `json:"action"`
	Target         string `json:"target"`
	Correlation_id string `json:"correlation_id"`
	Timestamp      string `json:"timetamp"`
	Control        string `json:"control"`
	Config         string `josn:"config"`
}
type Result struct {
	Latency string `json:"latency"`
	
}
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type Response struct {
	Type           string `json:"type"`
	Action         string `json:"action"`
	Config         string `json:"config"`
	Target         string `json:"target"`
	Status         string `json:"status"`
	Result         Result `json:"result"`
	Correlation_id string `json:"correlation"`
	Timestamp      string `json:"timetamp"`
	Control        string `json:"control"`
	Error          Error  `json:"error"`
}
type RunResult struct {
	Action        string `json:"action"`
	Status        string `json:"status"`
	Corrlation_id string `json:"correlation_id"`
	Timestamp     string `json:"timestamp"`
}
