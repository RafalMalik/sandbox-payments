package sdk

type ServiceOptions struct {
	Token   string
	BaseURL string
}

type Option func(*Client)

func WithRoundTripper(rt RoundTripper) Option {
	return func(c *Client) {
		c.exec.trip = rt
	}
}

type Client struct {
	configs map[string]ServiceOptions
	exec    Executor
}

func Configure(services map[string]ServiceOptions, opts ...Option) *Client {
	cp := map[string]ServiceOptions{}
	for k, v := range services {
		cp[k] = v
	}
	c := &Client{configs: cp}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) require(slug string) ServiceConfig {
	return makeService(slug, c.configs[slug])
}

func (c *Client) Tracker() *Tracker {
	return &Tracker{exec: &c.exec, service: c.require("tracker")}
}

func (c *Client) Yaman() *Yaman {
	return &Yaman{exec: &c.exec, service: c.require("yaman")}
}

func (c *Client) Todo() *Todo {
	return &Todo{live{&c.exec, c.require("todo")}}
}

func (c *Client) Desk() *Desk {
	return &Desk{live{&c.exec, c.require("desk")}}
}

func (c *Client) Scheduler() *Scheduler {
	return &Scheduler{live{&c.exec, c.require("scheduler")}}
}

func (c *Client) Pulse() *Pulse {
	return &Pulse{live{&c.exec, c.require("pulse")}}
}

func (c *Client) Postern() *Postern {
	return &Postern{live{&c.exec, c.require("postern")}}
}

func (c *Client) Invest() *Invest {
	return &Invest{live{&c.exec, c.require("invest")}}
}

func (c *Client) Abiso() *Abiso {
	return &Abiso{live{&c.exec, c.require("abiso")}}
}

func (c *Client) Bisita() *Bisita {
	return &Bisita{live{&c.exec, c.require("bisita")}}
}

func (c *Client) Tala() *Tala {
	return &Tala{live{&c.exec, c.require("tala")}}
}

func (c *Client) Aklat() *Aklat {
	return &Aklat{live{&c.exec, c.require("aklat")}}
}

func (c *Client) Flow() *Flow {
	return &Flow{live{&c.exec, c.require("flow")}}
}

func (c *Client) Perfecto() *Perfecto {
	return &Perfecto{live{&c.exec, c.require("perfecto")}}
}

func (c *Client) Verdict() *Verdict {
	return &Verdict{live{&c.exec, c.require("verdict")}}
}

func (c *Client) Hanap() *Hanap {
	return &Hanap{live{&c.exec, c.require("hanap")}}
}

func (c *Client) Storage() *Storage {
	return &Storage{exec: &c.exec, service: c.require("storage")}
}

func (c *Client) PolicyManager() *PolicyManager {
	return &PolicyManager{exec: &c.exec, service: c.require("policy-manager")}
}

func (c *Client) SandboxPayments() *SandboxPayments {
	return &SandboxPayments{exec: &c.exec, service: c.require("sandbox-payments")}
}

func (c *Client) Bantay() *Bantay {
	return &Bantay{exec: &c.exec, service: c.require("bantay")}
}

func (c *Client) AiCenter() *AiCenter {
	return &AiCenter{exec: &c.exec, service: c.require("ai-center")}
}

func (c *Client) DatasetBuilder() *DatasetBuilder {
	return &DatasetBuilder{exec: &c.exec, service: c.require("dataset-builder")}
}

func (c *Client) Service(slug string) *GenericService {
	return &GenericService{exec: &c.exec, service: c.require(slug)}
}

type GenericService struct {
	exec    *Executor
	service ServiceConfig
}

func (g *GenericService) Request(method, path string, query map[string]string, jsonBody any) (map[string]any, error) {
	return g.exec.Call(g.service, method, path, query, jsonBody, true)
}

func (g *GenericService) Health() (map[string]any, error) {
	return g.exec.Call(g.service, "GET", HealthPath(g.service.Slug), nil, nil, false)
}

func (g *GenericService) Version() (map[string]any, error) {
	return g.exec.Call(g.service, "GET", "/version", nil, nil, false)
}
