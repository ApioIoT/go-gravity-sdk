package gravityworker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Worker struct {
	url string
}

func New(url string) Worker {
	return Worker{
		url: url,
	}
}

func (w *Worker) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, w.url, nil)
	if err != nil {
		return err
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return errors.New("gravity worker: can't connect to Apio Gravity")
	}

	return nil
}

func (w *Worker) Topic(name string, createTopic bool) (*Topic, error) {
	if createTopic {
		body, err := json.Marshal(TopicRequest{Uuid: name})
		if err != nil {
			return nil, err
		}

		u, err := url.JoinPath(w.url, "topics")
		if err != nil {
			return nil, err
		}

		resp, err := http.Post(u, "application/json", bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			var apioErr ApioError

			if err := json.Unmarshal(b, &apioErr); err != nil {
				return nil, err
			}

			return nil, errors.New(apioErr.Message)
		}
	}

	return &Topic{
		name:       name,
		gravityUrl: w.url,
	}, nil
}

func (w *Worker) Complete(job *Job, out interface{}) error {
	if job == nil {
		return errors.New("gravity worker: can't complete null job")
	}

	u, err := url.JoinPath(w.url, "jobs", job.Uuid, "complete")
	if err != nil {
		return err
	}

	var body JobRequest
	body.Output = out

	jBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, u, bytes.NewReader(jBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return errors.New("gravity worker: can't complete job")
	}

	return nil
}

func (w *Worker) Fail(job *Job, customError interface{}) error {
	if job == nil {
		return errors.New("gravity worker: can't fail null job")
	}

	u, err := url.JoinPath(w.url, "jobs", job.Uuid, "fail")
	if err != nil {
		return err
	}

	var body JobRequest
	body.Error = customError

	jBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, u, bytes.NewReader(jBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return errors.New("gravity worker: can't fail job")
	}

	return nil
}

func (w *Worker) Return(job *Job) error {
	if job == nil {
		return errors.New("gravity worker: can't return null job")
	}

	u, err := url.JoinPath(w.url, "jobs", job.Uuid, "return")
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, u, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return errors.New("gravity worker: can't return job")
	}

	return nil
}
