package gravityworker

import (
	"bytes"
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

type Topic struct {
	name       string
	gravityUrl string
}

func (t *Topic) Enqueue(payload interface{}) (*Job, error) {
	bPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	u, err := url.JoinPath(t.gravityUrl, "topics", t.name, "enqueue")
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(u, "application/json", bytes.NewReader(bPayload))
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

func (t *Topic) Dequeue() (*Job, error) {
	u, err := url.JoinPath(t.gravityUrl, "topics", t.name, "dequeue")
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

func (t *Topic) Listen(scheduling string, timezone string) (chan *Job, func() error, error) {
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
		job, err := t.Dequeue()
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
