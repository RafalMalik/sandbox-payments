package sdk

import "net/url"

type live struct {
	exec    *Executor
	service ServiceConfig
}

func (l live) Health() (map[string]any, error) {
	return l.exec.Call(l.service, "GET", HealthPath(l.service.Slug), nil, nil, false)
}

func (l live) Version() (map[string]any, error) {
	return l.exec.Call(l.service, "GET", "/version", nil, nil, false)
}

type Todo struct{ live }

func (t *Todo) Projects() *TodoProjects { return &TodoProjects{t.exec, t.service} }
func (t *Todo) Items() *TodoItems       { return &TodoItems{t.exec, t.service} }
func (t *Todo) ChecklistMatrix() (map[string]any, error) {
	return t.exec.Call(t.service, "GET", "/api/v1/checklist-matrix", nil, nil, true)
}
func (t *Todo) Queue() *PathList    { return &PathList{t.exec, t.service, "/api/v1/queue"} }
func (t *Todo) Activity() *PathList { return &PathList{t.exec, t.service, "/api/v1/activity"} }
func (t *Todo) Approved() *PathList { return &PathList{t.exec, t.service, "/api/v1/approved"} }
func (t *Todo) Links() *TodoLinks   { return &TodoLinks{t.exec, t.service} }

type PathList struct {
	exec    *Executor
	service ServiceConfig
	path    string
}

func (p *PathList) List() (map[string]any, error) {
	return p.exec.Call(p.service, "GET", p.path, nil, nil, true)
}

type TodoProjects struct {
	exec    *Executor
	service ServiceConfig
}

func (p *TodoProjects) List(query map[string]string) (map[string]any, error) {
	return p.exec.Call(p.service, "GET", "/api/v1/projects", query, nil, true)
}
func (p *TodoProjects) Get(slug string) (map[string]any, error) {
	return p.exec.Call(p.service, "GET", "/api/v1/projects/"+url.PathEscape(slug), nil, nil, true)
}
func (p *TodoProjects) Create(body map[string]any) (map[string]any, error) {
	return p.exec.Call(p.service, "POST", "/api/v1/projects", nil, body, true)
}

type TodoItems struct {
	exec    *Executor
	service ServiceConfig
}

func (p *TodoItems) List(projectSlug string, query map[string]string) (map[string]any, error) {
	return p.exec.Call(p.service, "GET", "/api/v1/projects/"+url.PathEscape(projectSlug)+"/items", query, nil, true)
}
func (p *TodoItems) Get(projectSlug, id string) (map[string]any, error) {
	return p.exec.Call(p.service, "GET", "/api/v1/projects/"+url.PathEscape(projectSlug)+"/items/"+url.PathEscape(id), nil, nil, true)
}
func (p *TodoItems) Create(projectSlug string, body map[string]any) (map[string]any, error) {
	return p.exec.Call(p.service, "POST", "/api/v1/projects/"+url.PathEscape(projectSlug)+"/items", nil, body, true)
}

type TodoLinks struct {
	exec    *Executor
	service ServiceConfig
}

func (p *TodoLinks) Lookup(query map[string]string) (map[string]any, error) {
	return p.exec.Call(p.service, "GET", "/api/v1/links", query, nil, true)
}

type Desk struct{ live }

func (d *Desk) Healthz() (map[string]any, error) {
	return d.exec.Call(d.service, "GET", "/healthz", nil, nil, false)
}
func (d *Desk) Projects() *DeskProjects { return &DeskProjects{d.exec, d.service} }
func (d *Desk) SpendPrompt() *SpendPromptAPI {
	return &SpendPromptAPI{d.exec, d.service}
}

type DeskProjects struct {
	exec    *Executor
	service ServiceConfig
}

func (p *DeskProjects) List(query map[string]string) (map[string]any, error) {
	return p.exec.Call(p.service, "GET", "/api/v1/projects", query, nil, true)
}
func (p *DeskProjects) Create(body map[string]any) (map[string]any, error) {
	return p.exec.Call(p.service, "POST", "/api/v1/projects", nil, body, true)
}

type SpendPromptAPI struct {
	exec    *Executor
	service ServiceConfig
}

func (p *SpendPromptAPI) Get() (map[string]any, error) {
	return p.exec.Call(p.service, "GET", "/api/v1/invest/spend-prompt", nil, nil, true)
}
func (p *SpendPromptAPI) Submit(body map[string]any) (map[string]any, error) {
	return p.exec.Call(p.service, "POST", "/api/v1/invest/spend-prompt", nil, body, true)
}

type Scheduler struct{ live }

func (s *Scheduler) Jobs() *IDCollection { return &IDCollection{s.exec, s.service, "/api/v1/jobs"} }

type IDCollection struct {
	exec    *Executor
	service ServiceConfig
	prefix  string
}

func (c *IDCollection) List() (map[string]any, error) {
	return c.exec.Call(c.service, "GET", c.prefix, nil, nil, true)
}
func (c *IDCollection) Get(id string) (map[string]any, error) {
	return c.exec.Call(c.service, "GET", c.prefix+"/"+url.PathEscape(id), nil, nil, true)
}
func (c *IDCollection) Create(body map[string]any) (map[string]any, error) {
	return c.exec.Call(c.service, "POST", c.prefix, nil, body, true)
}
func (c *IDCollection) Update(id string, body map[string]any) (map[string]any, error) {
	return c.exec.Call(c.service, "PATCH", c.prefix+"/"+url.PathEscape(id), nil, body, true)
}
func (c *IDCollection) Delete(id string) (map[string]any, error) {
	return c.exec.Call(c.service, "DELETE", c.prefix+"/"+url.PathEscape(id), nil, nil, true)
}

type Pulse struct{ live }

func (p *Pulse) Systems() *PulseSystems { return &PulseSystems{p.exec, p.service} }
func (p *Pulse) PublicBoard() *PulsePublic {
	return &PulsePublic{p.exec, p.service}
}

type PulseSystems struct {
	exec    *Executor
	service ServiceConfig
}

func (p *PulseSystems) List() (map[string]any, error) {
	return p.exec.Call(p.service, "GET", "/api/v1/systems", nil, nil, true)
}
func (p *PulseSystems) Get(slug string) (map[string]any, error) {
	return p.exec.Call(p.service, "GET", "/api/v1/systems/"+url.PathEscape(slug), nil, nil, true)
}

type PulsePublic struct {
	exec    *Executor
	service ServiceConfig
}

func (p *PulsePublic) Systems() (map[string]any, error) {
	return p.exec.Call(p.service, "GET", "/api/v1/public/systems", nil, nil, false)
}
func (p *PulsePublic) System(slug string) (map[string]any, error) {
	return p.exec.Call(p.service, "GET", "/api/v1/public/systems/"+url.PathEscape(slug), nil, nil, false)
}
func (p *PulsePublic) Map() (map[string]any, error) {
	return p.exec.Call(p.service, "GET", "/api/v1/public/map", nil, nil, false)
}

type Postern struct{ live }

func (p *Postern) Mailboxes() *IDCollection {
	return &IDCollection{p.exec, p.service, "/api/v1/mailboxes"}
}
func (p *Postern) Activity() *PathList { return &PathList{p.exec, p.service, "/api/v1/activity"} }

type Invest struct{ live }

func (i *Invest) Briefs() *PathList { return &PathList{i.exec, i.service, "/api/v1/briefs"} }
func (i *Invest) Spend() *InvestSpend {
	return &InvestSpend{i.exec, i.service}
}
func (i *Invest) Holdings() *IDCollection {
	return &IDCollection{i.exec, i.service, "/api/v1/holdings"}
}

type InvestSpend struct {
	exec    *Executor
	service ServiceConfig
}

func (s *InvestSpend) Prompt() (map[string]any, error) {
	return s.exec.Call(s.service, "GET", "/api/v1/spend-prompt", nil, nil, true)
}
func (s *InvestSpend) Submit(body map[string]any) (map[string]any, error) {
	return s.exec.Call(s.service, "POST", "/api/v1/spend-prompt", nil, body, true)
}
func (s *InvestSpend) History() (map[string]any, error) {
	return s.exec.Call(s.service, "GET", "/api/v1/spend-history", nil, nil, true)
}

type Abiso struct{ live }

func (a *Abiso) Products() *AbisoProducts { return &AbisoProducts{a.exec, a.service} }
func (a *Abiso) Events() *AbisoEvents    { return &AbisoEvents{a.exec, a.service} }
func (a *Abiso) Actions() *PathList      { return &PathList{a.exec, a.service, "/api/v1/actions"} }
func (a *Abiso) Deliveries() *PathList   { return &PathList{a.exec, a.service, "/api/v1/deliveries"} }
func (a *Abiso) Notifications() *AbisoNotifications {
	return &AbisoNotifications{a.exec, a.service}
}

type AbisoProducts struct {
	exec    *Executor
	service ServiceConfig
}

func (p *AbisoProducts) List() (map[string]any, error) {
	return p.exec.Call(p.service, "GET", "/api/v1/products", nil, nil, true)
}
func (p *AbisoProducts) Theme(product string) (map[string]any, error) {
	return p.exec.Call(p.service, "GET", "/api/v1/products/"+url.PathEscape(product)+"/theme", nil, nil, true)
}
func (p *AbisoProducts) PutTheme(product string, body map[string]any) (map[string]any, error) {
	return p.exec.Call(p.service, "PUT", "/api/v1/products/"+url.PathEscape(product)+"/theme", nil, body, true)
}

type AbisoEvents struct {
	exec    *Executor
	service ServiceConfig
}

func (p *AbisoEvents) Create(body map[string]any) (map[string]any, error) {
	return p.exec.Call(p.service, "POST", "/api/v1/events", nil, body, true)
}

type AbisoNotifications struct {
	exec    *Executor
	service ServiceConfig
}

func (p *AbisoNotifications) List(query map[string]string) (map[string]any, error) {
	return p.exec.Call(p.service, "GET", "/api/v1/notifications", query, nil, true)
}

type Bisita struct{ live }

func (b *Bisita) Visits() *BisitaVisits   { return &BisitaVisits{b.exec, b.service} }
func (b *Bisita) Summary() *BisitaSummary { return &BisitaSummary{b.exec, b.service} }

type BisitaVisits struct {
	exec    *Executor
	service ServiceConfig
}

func (p *BisitaVisits) Create(body map[string]any) (map[string]any, error) {
	return p.exec.Call(p.service, "POST", "/api/v1/visits", nil, body, true)
}

type BisitaSummary struct {
	exec    *Executor
	service ServiceConfig
}

func (p *BisitaSummary) Get() (map[string]any, error) {
	return p.exec.Call(p.service, "GET", "/api/v1/summary", nil, nil, true)
}

type Tala struct{ live }

func (t *Tala) Calendars() *IDCollection { return &IDCollection{t.exec, t.service, "/api/v1/calendars"} }
func (t *Tala) Items() *IDCollection    { return &IDCollection{t.exec, t.service, "/api/v1/items"} }

type Aklat struct{ live }

func (a *Aklat) Builds() *AklatBuilds { return &AklatBuilds{a.exec, a.service} }

type AklatBuilds struct {
	exec    *Executor
	service ServiceConfig
}

func (p *AklatBuilds) Create(body map[string]any) (map[string]any, error) {
	return p.exec.Call(p.service, "POST", "/api/v1/builds", nil, body, true)
}

type Flow struct{ live }

func (f *Flow) Flows() *IDCollection { return &IDCollection{f.exec, f.service, "/api/v1/flows"} }
func (f *Flow) Runs() *FlowRuns     { return &FlowRuns{f.exec, f.service} }

type FlowRuns struct {
	exec    *Executor
	service ServiceConfig
}

func (p *FlowRuns) Get(id string) (map[string]any, error) {
	return p.exec.Call(p.service, "GET", "/api/v1/runs/"+url.PathEscape(id), nil, nil, true)
}
func (p *FlowRuns) Create(flowID string, body map[string]any) (map[string]any, error) {
	if body == nil {
		body = map[string]any{}
	}
	return p.exec.Call(p.service, "POST", "/api/v1/flows/"+url.PathEscape(flowID)+"/runs", nil, body, true)
}

type Perfecto struct{ live }

func (p *Perfecto) Projects() *PerfectoProjects { return &PerfectoProjects{p.exec, p.service} }
func (p *Perfecto) Keys() *IDCollection        { return &IDCollection{p.exec, p.service, "/api/v1/keys"} }

type PerfectoProjects struct {
	exec    *Executor
	service ServiceConfig
}

func (p *PerfectoProjects) List() (map[string]any, error) {
	return p.exec.Call(p.service, "GET", "/api/v1/projects", nil, nil, true)
}
func (p *PerfectoProjects) Create(body map[string]any) (map[string]any, error) {
	return p.exec.Call(p.service, "POST", "/api/v1/projects", nil, body, true)
}
func (p *PerfectoProjects) CreateRun(slug string, body map[string]any) (map[string]any, error) {
	return p.exec.Call(p.service, "POST", "/api/v1/projects/"+url.PathEscape(slug)+"/runs", nil, body, true)
}

type Verdict struct{ live }

type Storage struct {
	exec    *Executor
	service ServiceConfig
}

func (s *Storage) Health() (map[string]any, error) {
	return s.exec.Call(s.service, "GET", HealthPath(s.service.Slug), nil, nil, false)
}
func (s *Storage) Objects() *StorageObjects { return &StorageObjects{s.exec, s.service} }

type StorageObjects struct {
	exec    *Executor
	service ServiceConfig
}

func (p *StorageObjects) List(query map[string]string) (map[string]any, error) {
	return p.exec.Call(p.service, "GET", "/api/v1/objects", query, nil, true)
}

type PolicyManager struct {
	exec    *Executor
	service ServiceConfig
}

func (p *PolicyManager) Health() (map[string]any, error) {
	return p.exec.Call(p.service, "GET", HealthPath(p.service.Slug), nil, nil, false)
}
func (p *PolicyManager) Check(body map[string]any) (map[string]any, error) {
	return p.exec.Call(p.service, "POST", "/api/v1/check", nil, body, true)
}
func (p *PolicyManager) Use(body map[string]any) (map[string]any, error) {
	return p.exec.Call(p.service, "POST", "/api/v1/use", nil, body, true)
}

type Hanap struct{ live }

func (h *Hanap) Ready() (map[string]any, error) {
	return h.exec.Call(h.service, "GET", "/ready", nil, nil, false)
}
func (h *Hanap) Search(body map[string]any) (map[string]any, error) {
	return h.exec.Call(h.service, "POST", "/v1/search", nil, body, true)
}

type SandboxPayments struct {
	exec    *Executor
	service ServiceConfig
}

func (s *SandboxPayments) Health() (map[string]any, error) {
	return s.exec.Call(s.service, "GET", HealthPath(s.service.Slug), nil, nil, false)
}
func (s *SandboxPayments) PaymentMethods() *PathList {
	return &PathList{s.exec, s.service, "/api/payment-methods"}
}
func (s *SandboxPayments) Payments() *IDCollection {
	return &IDCollection{s.exec, s.service, "/api/payments"}
}

type Bantay struct {
	exec    *Executor
	service ServiceConfig
}

func (b *Bantay) Health() (map[string]any, error) {
	return b.exec.Call(b.service, "GET", HealthPath(b.service.Slug), nil, nil, false)
}

type AiCenter struct {
	exec    *Executor
	service ServiceConfig
}

func (a *AiCenter) Health() (map[string]any, error) {
	return a.exec.Call(a.service, "GET", HealthPath(a.service.Slug), nil, nil, false)
}
func (a *AiCenter) Version() (map[string]any, error) {
	return a.exec.Call(a.service, "GET", "/api/version", nil, nil, false)
}

type DatasetBuilder struct {
	exec    *Executor
	service ServiceConfig
}

func (d *DatasetBuilder) Health() (map[string]any, error) {
	return d.exec.Call(d.service, "GET", HealthPath(d.service.Slug), nil, nil, false)
}
