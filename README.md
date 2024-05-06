### Go Gravity Worker Helper  

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

#### Enqueue job
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

#### Listen jobs
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
  // For complete a job
  if err := gravity.Complete(job, nil); err != nil {
    log.Fatal(err)
  } 
  
  // For fail a job
  if err := gravity.Fail(job, nil); err != nil {
    log.Fatal(err)
  } 
  
  // For return a job
  if err := gravity.Return(job); err != nil {
    log.Fatal(err)
  } 
}
```