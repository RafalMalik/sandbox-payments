# github.com/raspberry-vision/sdk-go

Raspberry Vision platform HTTP SDK. Typed clients for Tracker, Yaman, Todo, Desk, and the rest of the product catalog.

**Server-side only.** Do not put an API token in a browser bundle.

```go
rv := sdk.Configure(map[string]sdk.ServiceOptions{
    "tracker": {Token: os.Getenv("TRACKER_API_TOKEN")},
    "todo":    {Token: os.Getenv("TODO_API_TOKEN")},
})
_, err := rv.Tracker().NotifyDeploy(map[string]any{
    "system":  "desk",
    "version": "1.2.0",
    "status":  "ok",
})
_, err = rv.Yaman().Secrets().Get("panel-key")
_, err = rv.Todo().Projects().List(map[string]string{"status": "active"})
_, err = rv.PolicyManager().Check(map[string]any{"action": "queue.submit"})
```

```bash
go get github.com/raspberry-vision/sdk-go@v0.3.0
```

See [`docs/API.md`](docs/API.md).
