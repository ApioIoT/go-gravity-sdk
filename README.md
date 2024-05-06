## Go Gravity Worker Helper  

Helps you implement gravity job workers:

#### Install
```bash
go get 
```

#### Test Connection
```golang
gravity := gravityworker.New("http://localhost:7000")

if err := gravity.Ping(); err != nil {
  log.Fatal(err)
}
```

#### Enqueue a job
```golang
type Payload struct {
  Message string `json:"message"`
}

gravity := gravityworker.New("http://localhost:7000")

topic, err := gravity.Topic("project.resource.action", true)
if err != nil {
  log.Fatal(err)
}

if err := topic.Enqueue(Payload{ Message: "ciao" }); err != nil {
  log.Fatal(err)
}
```

#### Listen for a job
```golang
gravity := gravityworker.New("http://localhost:7000")

topic, err := gravity.Topic("project.resource.action", true)
if err != nil {
  log.Fatal(err)
}

jobs, cancel, err := topic.Listen("* * * * * *", "Europe/Rome")
if err != nil {
  log.Fatal(err)
}
defer func() {
  if err := cancel(); err != nil {
    log.Fatal(err)
  }
}()

for job := range jobs {
  // ...
}
```

#### Complete, Fail and Return
```golang
gravity := gravityworker.New("http://localhost:7000")

topic, err := gravity.Topic("project.resource.action", true)
if err != nil {
  log.Fatal(err)
}

job, err := topic.Dequeue()
if err != nil {
  t.Fatal(err)
}

// For Complete
if err := gravity.Complete(job, nil); err != nil {
  log.Fatal(err)
} 

// For Fail
if err := gravity.Fail(job, nil); err != nil {
  log.Fatal(err)
} 

// For Return
if err := gravity.Return(job); err != nil {
  log.Fatal(err)
} 

```
