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
