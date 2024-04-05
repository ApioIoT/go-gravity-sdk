package apio

type ApioResponse[T any] struct {
	Status bool `json:"status,omitempty"`
	Data   T    `json:"data,omitempty"`
}

type ApioResponseError struct {
	Status bool `json:"status,omitempty"`
	Error  struct {
		Name       string `json:"name,omitempty"`
		StatusCode int    `json:"statusCode"`
		Message    string `json:"message"`
		Code       string `json:"code,omitempty"`
	} `json:"error,omitempty"`
}

type JobRequest struct {
	Error  interface{} `json:"error,omitempty"`
	Output interface{} `json:"out,omitempty"`
}
