package test

import (
	"apio/go-gravity-worker/pkg/gravityworker"
	"strings"
	"testing"
)

const (
	GRAVITY_URL   = "http://localhost:10005"
	GRAVITY_TOPIC = "first-topic"
)

func TestConnection(t *testing.T) {
	worker := gravityworker.New(GRAVITY_URL)

	if err := worker.Ping(); err != nil {
		t.Fatal(err)
	}
}

func TestEnqueue(t *testing.T) {
	worker := gravityworker.New(GRAVITY_URL)

	topic, err := worker.Topic(GRAVITY_TOPIC, true)
	if err != nil {
		t.Fatal(err)
	}

	type payload struct {
		Message string `json:"message"`
	}

	if _, err := topic.Enqueue(payload{Message: "Job for complete"}); err != nil {
		t.Fatal(err)
	}
	if _, err := topic.Enqueue(payload{Message: "Job for fail"}); err != nil {
		t.Fatal(err)
	}
	if _, err := topic.Enqueue(payload{Message: "Job for read"}); err != nil {
		t.Fatal(err)
	}
}

func TestComplete(t *testing.T) {
	worker := gravityworker.New(GRAVITY_URL)

	topic, err := worker.Topic(GRAVITY_TOPIC, true)
	if err != nil {
		t.Fatal(err)
	}

	job, err := topic.Dequeue()
	if err != nil {
		t.Fatal(err)
	}

	if err := worker.Complete(job, nil); err != nil {
		t.Fatal(err)
	}
}

func TestFail(t *testing.T) {
	worker := gravityworker.New(GRAVITY_URL)

	topic, err := worker.Topic(GRAVITY_TOPIC, true)
	if err != nil {
		t.Fatal(err)
	}

	job, err := topic.Dequeue()
	if err != nil {
		t.Fatal(err)
	}

	if err := worker.Fail(job, nil); err != nil {
		t.Fatal(err)
	}
}

func TestRead(t *testing.T) {
	worker := gravityworker.New(GRAVITY_URL)

	topic, err := worker.Topic(GRAVITY_TOPIC, true)
	if err != nil {
		t.Fatal(err)
	}

	job, err := topic.Dequeue()
	if err != nil {
		t.Fatal(err)
	}

	data, ok := job.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Can't cast response")
	}

	message, ok := data["message"].(string)
	if !ok {
		t.Fatal("Can't cast response")
	}

	if !strings.HasPrefix(message, "Job") {
		t.Fatal("Invalid read data")
	}
}
