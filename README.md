### Go Gravity Worker Helper  

Helps you implement gravity job workers:

#### Listen jobs
```golang
worker := gravityworker.New("project.resource.action", "http://gravity:7000", "* * * * * *", "Europe/Rome")

if err := worker.Start(); err != nil {
  log.Fatal(err)
}
defer worker.Stop()

for job := range worker.Jobs() {
  job.Complete(nil) // For complete a job
  job.Fail(nil)     // For fail a job
  job.Return()      // For return a job
}
```

#### Enqueue job
```golang
type Payload struct {
	Message string `json:"message"`
}

worker := gravityworker.New("project.resource.action", "http://gravity:7000", "* * * * *", "Europe/Rome")
worker.Enqueue(Payload{
  Message: "ciao",
})
```
