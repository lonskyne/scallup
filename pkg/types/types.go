package types

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type NotLeaderError struct {
    LeaderID int
}

func (e *NotLeaderError) Error() string {
    return "request received by a node that is not the leader"
}
