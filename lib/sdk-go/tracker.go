package sdk

type Tracker struct {
	exec    *Executor
	service ServiceConfig
}

func (t *Tracker) Health() (map[string]any, error) {
	return t.exec.Call(t.service, "GET", HealthPath(t.service.Slug), nil, nil, false)
}

func (t *Tracker) Version() (map[string]any, error) {
	return t.exec.Call(t.service, "GET", "/version", nil, nil, false)
}

func (t *Tracker) Events() *EventsAPI {
	return &EventsAPI{exec: t.exec, service: t.service}
}

func (t *Tracker) Deployments() *DeploymentsAPI {
	return &DeploymentsAPI{exec: t.exec, service: t.service}
}

func (t *Tracker) Systems() *SystemsAPI {
	return &SystemsAPI{exec: t.exec, service: t.service}
}

func (t *Tracker) Completeness() *CompletenessAPI {
	return &CompletenessAPI{exec: t.exec, service: t.service}
}

func (t *Tracker) NotifyDeploy(body map[string]any) (map[string]any, error) {
	payload := map[string]any{}
	for k, v := range body {
		payload[k] = v
	}
	if src, _ := payload["source"].(string); src == "" {
		payload["source"] = "github-actions"
	}
	return t.Deployments().Create(payload)
}

type EventsAPI struct {
	exec    *Executor
	service ServiceConfig
}

func (e *EventsAPI) Create(body map[string]any) (map[string]any, error) {
	return e.exec.Call(e.service, "POST", "/api/v1/events", nil, body, true)
}

func (e *EventsAPI) List(filters map[string]string) (map[string]any, error) {
	return e.exec.Call(e.service, "GET", "/api/v1/events", filters, nil, true)
}

type DeploymentsAPI struct {
	exec    *Executor
	service ServiceConfig
}

func (d *DeploymentsAPI) Create(body map[string]any) (map[string]any, error) {
	return d.exec.Call(d.service, "POST", "/api/v1/deployments", nil, body, true)
}

type SystemsAPI struct {
	exec    *Executor
	service ServiceConfig
}

func (s *SystemsAPI) List() (map[string]any, error) {
	return s.exec.Call(s.service, "GET", "/api/v1/systems", nil, nil, true)
}

func (s *SystemsAPI) Get(slug string) (map[string]any, error) {
	return s.exec.Call(s.service, "GET", "/api/v1/systems/"+slug, nil, nil, true)
}

func (s *SystemsAPI) Create(body map[string]any) (map[string]any, error) {
	return s.exec.Call(s.service, "POST", "/api/v1/systems", nil, body, true)
}

func (s *SystemsAPI) Update(slug string, body map[string]any) (map[string]any, error) {
	return s.exec.Call(s.service, "PUT", "/api/v1/systems/"+slug, nil, body, true)
}

func (s *SystemsAPI) Delete(slug string) (map[string]any, error) {
	return s.exec.Call(s.service, "DELETE", "/api/v1/systems/"+slug, nil, nil, true)
}

func (s *SystemsAPI) SetVersion(slug, version string) (map[string]any, error) {
	return s.exec.Call(s.service, "POST", "/api/v1/systems/"+slug+"/version", nil, map[string]any{"version": version}, true)
}

type CompletenessAPI struct {
	exec    *Executor
	service ServiceConfig
}

func (c *CompletenessAPI) Ingest(body map[string]any) (map[string]any, error) {
	return c.exec.Call(c.service, "POST", "/api/v1/completeness", nil, body, true)
}
