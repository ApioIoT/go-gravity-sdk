### Go Gravity Worker Helper  

Helps you implement gravity job workers:

#### Listen jobs
```golang
worker := gravityworker.New("project.resource.action", "http://gravity:7000", "* * * * *", "Europe/Rome")

if err := w.Start(); err != nil {
  t.Fatal(err)
}

for job := range worker.Jobs() {
  job.Complete(nil)
  job.Fail(nil)
  job.Return()
}

worker.Stop()
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
