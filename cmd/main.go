package main

import (
	gravityworker "apio/go-gravity-worker/pkg/worker"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
)

func main() {
	w := gravityworker.New("first-topic", "http://localhost:10005", "* * * * * *", "Europe/Rome")

	if err := w.Start(); err != nil {
		log.Fatal(err)
	}
	defer w.Stop()

	for j := range w.Jobs() {
		data, err := json.Marshal(j)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(data))

		n := rand.Intn(2)

		if n == 0 {
			if err := j.Complete(nil); err != nil {
				fmt.Println(err)
			}
		} else if n == 1 {
			if err := j.Fail(nil); err != nil {
				fmt.Println(err)
			}
		} else {
			if err := j.Complete(nil); err != nil {
				fmt.Println(err)
			}
		}
	}
}
