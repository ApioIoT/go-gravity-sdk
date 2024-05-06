### Go Gravity Worker Helper  

Helps you implement gravity job workers:

#### Install
```bash
  go get 
```

#### Listen jobs
```golang
worker, err := gravityworker.New("http://localhost:10005", "project.resource.action", true)
if err != nil {
  log.Fatal(err)
}

jobs, cancel, err := worker.Listen("* * * * * *", "Europe/Rome")
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
  if err := worker.Complete(job, nil); err != nil {
    log.Fatal(err)
  } 
  
  // For fail a job
  if err := worker.Fail(job, nil); err != nil {
    log.Fatal(err)
  } 
  
  // For return a job
  if err := worker.Return(job); err != nil {
    log.Fatal(err)
  } 
}
```

#### Enqueue job
```golang
type Payload struct {
  Message string `json:"message"`
}

worker, err := gravityworker.New("http://gravity:7000", "project.resource.action", true)
if err != nil {
  log.Fatal(err)
}

if err := worker.Enqueue(Payload{ Message: "ciao" }); err != nil {
  log.Fatal(err)
}
```
