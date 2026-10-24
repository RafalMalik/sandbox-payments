package sdk

import "net/url"

type Yaman struct {
	exec    *Executor
	service ServiceConfig
}

func (y *Yaman) Health() (map[string]any, error) {
	return y.exec.Call(y.service, "GET", HealthPath(y.service.Slug), nil, nil, false)
}

func (y *Yaman) Secrets() *SecretsAPI {
	return &SecretsAPI{exec: y.exec, service: y.service}
}

type SecretsAPI struct {
	exec    *Executor
	service ServiceConfig
}

func (s *SecretsAPI) List() (map[string]any, error) {
	return s.exec.Call(s.service, "GET", "/api/v1/secrets", nil, nil, true)
}

func (s *SecretsAPI) Get(name string) (map[string]any, error) {
	return s.exec.Call(s.service, "GET", "/api/v1/secrets/"+url.PathEscape(name), nil, nil, true)
}

func (s *SecretsAPI) Create(body map[string]any) (map[string]any, error) {
	return s.exec.Call(s.service, "POST", "/api/v1/secrets", nil, body, true)
}

func (s *SecretsAPI) Update(name string, body map[string]any) (map[string]any, error) {
	return s.exec.Call(s.service, "PATCH", "/api/v1/secrets/"+url.PathEscape(name), nil, body, true)
}

func (s *SecretsAPI) Delete(name string) (map[string]any, error) {
	return s.exec.Call(s.service, "DELETE", "/api/v1/secrets/"+url.PathEscape(name), nil, nil, true)
}
