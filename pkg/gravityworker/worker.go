package gravityworker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-co-op/gocron/v2"
)

type Worker struct {
	url   string
	topic string
}

func New(gravityUrl string, topic string, createTopic bool) (*Worker, error) {
	if createTopic {
		body, err := json.Marshal(TopicRequest{Uuid: topic})
		if err != nil {
			return nil, err
		}

		u, err := url.JoinPath(gravityUrl, "topics")
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

	return &Worker{
		url:   gravityUrl,
		topic: topic,
	}, nil
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

func (w *Worker) Listen(scheduling string, timezone string) (chan *Job, func() error, error) {
	jobsChan := make(chan *Job, 100)

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, dummyFunc, err
	}

	scheduler, err := gocron.NewScheduler(gocron.WithLocation(location))
	if err != nil {
		return nil, dummyFunc, err
	}

	task := func() {
		job, err := w.Dequeue()
		if err != nil {
			apioErr, ok := err.(*ApioError)
			if !ok || apioErr.StatusCode != 404 {
				fmt.Fprintln(os.Stderr, err)
			}
		} else {
			jobsChan <- job
		}
	}

	j, err := scheduler.NewJob(gocron.CronJob(scheduling, true), gocron.NewTask(task))
	if err != nil {
		return nil, dummyFunc, err
	}

	scheduler.Start()

	if err := j.RunNow(); err != nil {
		return nil, dummyFunc, err
	}

	return jobsChan, func() error {
		if err := scheduler.Shutdown(); err != nil {
			return err
		}

		close(jobsChan)

		return nil
	}, nil
}

func (w *Worker) Enqueue(payload interface{}) (*Job, error) {
	bPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	u, err := url.JoinPath(w.url, "topics", w.topic, "enqueue")
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(bPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, errors.New("gravity worker: error adding job")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apioResp JobResponse

	if err := json.Unmarshal(body, &apioResp); err != nil {
		return nil, err
	}

	return &apioResp.Data, nil
}

func (w *Worker) Dequeue() (*Job, error) {
	u, err := url.JoinPath(w.url, "topics", w.topic, "dequeue")
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(u, "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var apioResp ErrorResponse
		if err := json.Unmarshal(body, &apioResp); err != nil {
			return nil, NewApioError(resp.StatusCode, "gravity worker: Error getting job")
		}
		return nil, NewApioError(resp.StatusCode, fmt.Sprintf("gravity worker: %s", apioResp.Error.Message))
	}

	var apioResp JobResponse
	if err := json.Unmarshal(body, &apioResp); err != nil {
		return nil, err
	}

	return &apioResp.Data, nil
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
