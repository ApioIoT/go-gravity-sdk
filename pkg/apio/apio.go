package apio

type ApioResponse[T any] struct {
	Status string `json:"status,omitempty"`
	Data   T      `json:"data,omitempty"`
}

type ApioResponseError struct {
	Status bool `json:"status,omitempty"`
	Error  struct {
		Name    string `json:"name,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"error,omitempty"`
}

type JobRequest struct {
	Error  interface{} `json:"error,omitempty"`
	Output interface{} `json:"out,omitempty"`
}
