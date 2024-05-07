package gravitysdk

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

type gravity struct {
	url string
}

func New(url string) gravity {
	return gravity{
		url: url,
	}
}

func (g *gravity) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.url, nil)
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

func (g *gravity) Topic(name string, createTopic bool) (*topic, error) {
	if createTopic {
		body, err := json.Marshal(topicRequest{Uuid: name})
		if err != nil {
			return nil, err
		}

		u, err := url.JoinPath(g.url, "topics")
		if err != nil {
			return nil, err
		}

		resp, err := http.Post(u, "application/json", bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 && resp.StatusCode != 409 {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			var apioErr apioError

			if err := json.Unmarshal(b, &apioErr); err != nil {
				return nil, err
			}

			return nil, errors.New(apioErr.Message)
		}
	}

	return &topic{
		name:       name,
		gravityUrl: g.url,
	}, nil
}

func (g *gravity) Complete(job *job, out interface{}) error {
	if job == nil {
		return errors.New("gravity worker: can't complete null job")
	}

	if job.Status != IN_PROGRESS {
		return errors.New("gravity worker: can't complete a not in_progress job")
	}

	u, err := url.JoinPath(g.url, "jobs", job.Uuid, "complete")
	if err != nil {
		return err
	}

	var body jobRequest
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

	job.Status = COMPLETED

	return nil
}

func (g *gravity) Fail(job *job, customError interface{}) error {
	if job == nil {
		return errors.New("gravity worker: can't fail null job")
	}

	if job.Status != IN_PROGRESS {
		return errors.New("gravity worker: can't fail a not in_progress job")
	}

	u, err := url.JoinPath(g.url, "jobs", job.Uuid, "fail")
	if err != nil {
		return err
	}

	var body jobRequest
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

	job.Status = FAILED

	return nil
}

func (g *gravity) Return(job *job) error {
	if job == nil {
		return errors.New("gravity worker: can't return null job")
	}

	if job.Status != IN_PROGRESS {
		return errors.New("gravity worker: can't return a not in_progress job")
	}

	u, err := url.JoinPath(g.url, "jobs", job.Uuid, "return")
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

	job.Status = QUEUED

	return nil
}
