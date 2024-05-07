package gogravity

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

const JOBS_BUFFER_SIZE = 100

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

	if resp.StatusCode >= 400 {
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

func (t *topic) Listen(scheduling string, timezone string) (<-chan *job, func() error, error) {
	jobsChan := make(chan *job, JOBS_BUFFER_SIZE)

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

func (t *topic) AddSchedule(crontab string, timezone string, active bool, once bool, delay int32) error {
	u, err := url.JoinPath(t.gravityUrl, "schedules")
	if err != nil {
		return err
	}

	type payloadType struct {
		Topic        string `json:"topic"`
		Cron         string `json:"cron"`
		CronTimezone string `json:"cronTimezone"`
		Active       bool   `json:"active"`
		ScheduleOnce bool   `json:"scheduleOnce"`
		Delay        int32  `json:"delay"`
	}

	payload := payloadType{
		Topic:        t.name,
		Cron:         crontab,
		CronTimezone: timezone,
		Active:       active,
		ScheduleOnce: once,
		Delay:        delay,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(u, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var apioErr apioError

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(body, &apioErr); err != nil {
			return err
		}

		return errors.New(apioErr.Message)
	}

	return nil
}
