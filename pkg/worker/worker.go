package gravityworker

import (
	"apio/go-gravity-worker/pkg/apio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-co-op/gocron/v2"
)

const (
	QUEUED      = "queued"
	IN_PROGRESS = "in_progress"
	COMPLETED   = "completed"
	FAILED      = "failed"
	SKIPPED     = "skipped"
)

type Job struct {
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
	gravityUrl   string      `json:"-"`
}

func (j *Job) Complete(out interface{}) error {
	u, err := url.JoinPath(j.gravityUrl, "jobs", j.Uuid, "complete")
	if err != nil {
		return err
	}

	var body apio.JobRequest
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

	if resp.StatusCode != 200 {
		return errors.New("job: can't complete job")
	}

	return nil
}

func (j *Job) Fail(customError interface{}) error {
	u, err := url.JoinPath(j.gravityUrl, "jobs", j.Uuid, "fail")
	if err != nil {
		return err
	}

	var body apio.JobRequest
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

	if resp.StatusCode != 200 {
		return errors.New("job: can't fail job")
	}

	return nil
}

func (j *Job) Return() error {
	u, err := url.JoinPath(j.gravityUrl, "jobs", j.Uuid, "return")
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

	if resp.StatusCode != 200 {
		return errors.New("job: can't return job")
	}

	return nil
}

type Worker struct {
	topic             string
	gravityUrl        string
	gravityScheduling string
	timezone          string
	jobsChan          chan Job
	scheduler         gocron.Scheduler
	started           bool
	stopped           bool
}

func (w *Worker) Jobs() <-chan Job {
	return w.jobsChan
}

func (w *Worker) Start() error {
	if w.started {
		return nil
	}

	w.started = true

	w.jobsChan = make(chan Job, 100)

	location, err := time.LoadLocation(w.timezone)
	if err != nil {
		return err
	}

	s, err := gocron.NewScheduler(gocron.WithLocation(location))
	if err != nil {
		return err
	}

	w.scheduler = s

	dequeue := func(w *Worker) {
		u, err := url.JoinPath(w.gravityUrl, "topics", w.topic, "dequeue")
		if err != nil {
			fmt.Println(err)
			return
		}

		resp, err := http.Post(u, "application/json", nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		if resp.StatusCode == 404 {
			return
		}

		if resp.StatusCode != 200 {
			var apioResp apio.ApioResponseError
			if err := json.Unmarshal(body, &apioResp); err != nil {
				fmt.Println("worker: Error getting job")
			} else {
				fmt.Println("worker:", apioResp.Error.Message)
			}
			w.Stop()
			return
		}

		var apioResp apio.ApioResponse[Job]
		if err := json.Unmarshal(body, &apioResp); err != nil {
			fmt.Println(err)
			return
		}

		job := apioResp.Data
		job.gravityUrl = w.gravityUrl

		w.jobsChan <- job
	}

	_, _ = w.scheduler.NewJob(gocron.CronJob(w.gravityScheduling, true), gocron.NewTask(dequeue, w))
	w.scheduler.Start()

	return nil
}

func (w *Worker) Stop() {
	if w.stopped {
		return
	}
	w.stopped = true

	w.scheduler.Shutdown()
	w.scheduler = nil

	close(w.jobsChan)
}

func (w *Worker) Enqueue(payload interface{}) (Job, error) {
	bPayload, err := json.Marshal(payload)
	if err != nil {
		return Job{}, err
	}

	u, err := url.JoinPath(w.gravityUrl, "topics", w.topic, "enqueue")
	if err != nil {
		return Job{}, err
	}

	resp, err := http.Post(u, "application/json", bytes.NewReader(bPayload))
	if err != nil {
		return Job{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return Job{}, errors.New("worker: error adding job")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Job{}, err
	}

	var apioResp apio.ApioResponse[Job]

	if err := json.Unmarshal(body, &apioResp); err != nil {
		return Job{}, err
	}

	job := apioResp.Data
	job.gravityUrl = w.gravityUrl

	return job, nil
}

func New(topic string, gravityUrl string, gravityScheduling string, timezone string) Worker {
	return Worker{
		topic:             topic,
		gravityUrl:        gravityUrl,
		gravityScheduling: gravityScheduling,
		timezone:          timezone,
		stopped:           false,
		started:           false,
	}
}
