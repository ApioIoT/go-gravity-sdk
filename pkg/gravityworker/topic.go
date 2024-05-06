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

type topic struct {
	name       string
	gravityUrl string
}

func (t *topic) Enqueue(payload interface{}) (*job, error) {
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

	if resp.StatusCode != 200 && resp.StatusCode != 409 {
		return nil, errors.New("gravity worker: error adding job")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apioResp jobResponse

	if err := json.Unmarshal(body, &apioResp); err != nil {
		return nil, err
	}

	return &apioResp.Data, nil
}

func (t *topic) Dequeue() (*job, error) {
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
		var apioResp errorResponse
		if err := json.Unmarshal(body, &apioResp); err != nil {
			return nil, newApioError(resp.StatusCode, "gravity worker: Error getting job")
		}
		return nil, newApioError(resp.StatusCode, fmt.Sprintf("gravity worker: %s", apioResp.Error.Message))
	}

	var apioResp jobResponse
	if err := json.Unmarshal(body, &apioResp); err != nil {
		return nil, err
	}

	return &apioResp.Data, nil
}

func (t *topic) Listen(scheduling string, timezone string) (chan *job, func() error, error) {
	jobsChan := make(chan *job, 100)

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
			apioErr, ok := err.(*apioError)
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
