package gravitysdk

import (
	"errors"
	"time"
)

// Job Status
const (
	QUEUED      = "queued"
	IN_PROGRESS = "in_progress"
	COMPLETED   = "completed"
	FAILED      = "failed"
	SKIPPED     = "skipped"
)

type job struct {
	Uuid         string      `json:"uuid,omitempty"`
	Data         interface{} `json:"data,omitempty"`
	Retries      int         `json:"retries,omitempty"`
	Priority     int         `json:"priority,omitempty"`
	BackoffUntil string      `json:"backoffUntil,omitempty"`
	Topic        string      `json:"topic,omitempty"`
	Status       string      `json:"status,omitempty"`
	WorkflowId   string      `json:"workflowId,omitempty"`
	Output       interface{} `json:"output,omitempty"`
	Error        interface{} `json:"error,omitempty"`
	CompletedAt  time.Time   `json:"completedAt,omitempty"`
	StartedAt    time.Time   `json:"startedAt,omitempty"`
}

type jobResponse struct {
	Status bool `json:"status,omitempty"`
	Data   job  `json:"data,omitempty"`
}

type jobRequest struct {
	Error  interface{} `json:"error,omitempty"`
	Output interface{} `json:"out,omitempty"`
}

type topicRequest struct {
	Uuid string `json:"uuid"`
}

type errorResponse struct {
	Status bool `json:"status,omitempty"`
	Error  struct {
		Name       string `json:"name,omitempty"`
		StatusCode int    `json:"statusCode"`
		Message    string `json:"message"`
		Code       string `json:"code,omitempty"`
	} `json:"error,omitempty"`
}

type apioError struct {
	StatusCode int
	Message    string
}

func (e *apioError) Error() string {
	return e.Message
}

func newApioError(statusCode int, message string) *apioError {
	return &apioError{StatusCode: statusCode, Message: message}
}

func dummyFunc() error {
	return errors.New("gravity worker: called stop function on unstarted worker")
}
