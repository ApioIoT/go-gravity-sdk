package test

import (
	gravityworker "apio/go-gravity-worker/pkg/worker"
	"encoding/json"
	"fmt"
	"testing"
)

var gravityUrl string

func TestMain(m *testing.M) {
	gravityUrl = "http://localhost:10005"
}

func Test1(t *testing.T) {
	w := gravityworker.New("first", gravityUrl, "* * * * *", "Europe/Rome")

	if err := w.Start(); err != nil {
		t.Fatal(err)
	}

	counter := 0

	for j := range w.Jobs() {
		counter++
		if counter == 5 {
			w.Stop()
			return
		}

		fmt.Println("New job")

		data, err := json.Marshal(j)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(string(data))
	}
}
