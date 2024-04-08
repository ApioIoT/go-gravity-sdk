package gravityworker

type JobResponse struct {
	Status bool `json:"status,omitempty"`
	Data   Job  `json:"data,omitempty"`
}

type JobRequest struct {
	Error  interface{} `json:"error,omitempty"`
	Output interface{} `json:"out,omitempty"`
}

type ErrorResponse struct {
	Status bool `json:"status,omitempty"`
	Error  struct {
		Name       string `json:"name,omitempty"`
		StatusCode int    `json:"statusCode"`
		Message    string `json:"message"`
		Code       string `json:"code,omitempty"`
	} `json:"error,omitempty"`
}

type ApioError struct {
	StatusCode int
	Message    string
}

func (e *ApioError) Error() string {
	return e.Message
}

func NewApioError(statusCode int, message string) *ApioError {
	return &ApioError{StatusCode: statusCode, Message: message}
}
