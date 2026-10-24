package sdk

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func loadFixture(t *testing.T, rel string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("tests", "fixtures", rel))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]any{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func loadRaw(t *testing.T, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("tests", "fixtures", rel))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

type queuedCall struct {
	method  string
	url     string
	headers map[string]string
	body    *string
}

func queuedTrip(t *testing.T, bodies []string, statuses []int, calls *[]queuedCall) RoundTripper {
	t.Helper()
	i := 0
	return func(method, rawURL string, headers map[string]string, body *string) (int, string, error) {
		*calls = append(*calls, queuedCall{method: method, url: rawURL, headers: headers, body: body})
		if i >= len(bodies) {
			t.Fatal("queue empty")
		}
		status := statuses[i]
		b := bodies[i]
		i++
		return status, b, nil
	}
}

func TestAuthHeaders(t *testing.T) {
	opaque := loadFixture(t, "auth/opaque-headers.json")
	token := opaque["token"].(string)
	got := AuthHeaders(token, false)
	want := opaque["headers"].(map[string]any)
	if got["Authorization"] != want["Authorization"].(string) {
		t.Fatalf("opaque Authorization")
	}
	if got["X-Api-Key"] != want["X-Api-Key"].(string) {
		t.Fatalf("opaque X-Api-Key")
	}
	jwt := loadFixture(t, "auth/jwt-headers.json")
	jh := AuthHeaders(jwt["token"].(string), false)
	if _, ok := jh["X-Api-Key"]; ok {
		t.Fatal("jwt must omit X-Api-Key")
	}
}

func TestTrackerFlow(t *testing.T) {
	var calls []queuedCall
	rt := queuedTrip(t, []string{
		loadRaw(t, "tracker/health.json"),
		loadRaw(t, "tracker/event-create-response.json"),
		loadRaw(t, "tracker/event-list-response.json"),
		loadRaw(t, "tracker/deployment-create-response.json"),
		loadRaw(t, "tracker/system-list-response.json"),
		loadRaw(t, "tracker/completeness-ingest-response.json"),
		`{"projects":[]}`,
	}, []int{200, 201, 200, 201, 200, 201, 200}, &calls)
	token := "trk_0123456789abcdef0123456789abcdef"
	rv := Configure(map[string]ServiceOptions{
		"tracker": {Token: token, BaseURL: "https://tracker.test"},
		"todo":    {Token: "todo_abc"},
	}, WithRoundTripper(rt))
	health, err := rv.Tracker().Health()
	if err != nil || health["status"] != "ok" {
		t.Fatalf("health: %v %v", health, err)
	}
	if _, ok := calls[0].headers["Authorization"]; ok {
		t.Fatal("health must skip auth")
	}
	created, err := rv.Tracker().Events().Create(loadFixture(t, "tracker/event-create-request.json"))
	if err != nil || created["type"] != "org.note.unknown" {
		t.Fatalf("create: %v %v", created, err)
	}
	if calls[1].headers["X-Api-Key"] != token {
		t.Fatal("opaque key missing")
	}
	listed, err := rv.Tracker().Events().List(map[string]string{"system": "deploy-target"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(calls[2].url, "system=deploy-target") {
		t.Fatalf("list url %s", calls[2].url)
	}
	events := listed["events"].([]any)
	first := events[0].(map[string]any)
	if first["type"] != "deploy.started" {
		t.Fatalf("list type %v", first["type"])
	}
	if _, err := rv.Tracker().NotifyDeploy(map[string]any{"system": "deploy-target", "version": "2.0.0", "status": "ok"}); err != nil {
		t.Fatal(err)
	}
	var sent map[string]any
	_ = json.Unmarshal([]byte(*calls[3].body), &sent)
	if sent["source"] != "github-actions" {
		t.Fatalf("source %v", sent["source"])
	}
	systems, err := rv.Tracker().Systems().List()
	if err != nil {
		t.Fatal(err)
	}
	if systems["systems"].([]any)[0].(map[string]any)["slug"] != "parent-sys" {
		t.Fatal("systems")
	}
	ingested, err := rv.Tracker().Completeness().Ingest(loadFixture(t, "tracker/completeness-ingest-request.json"))
	if err != nil || ingested["system"] != "ibiyahe" {
		t.Fatalf("completeness %v %v", ingested, err)
	}
	if _, err := rv.Service("todo").Request("GET", "/api/v1/projects", nil, nil); err != nil {
		t.Fatal(err)
	}
	if calls[6].url != "https://todo.raspberry-vision.cloud/api/v1/projects" {
		t.Fatalf("todo url %s", calls[6].url)
	}
}

func TestYamanSecrets(t *testing.T) {
	var calls []queuedCall
	rt := queuedTrip(t, []string{
		loadRaw(t, "yaman/get-ok.json"),
		loadRaw(t, "yaman/create-ok.json"),
	}, []int{200, 201}, &calls)
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyIn0.sig"
	rv := Configure(map[string]ServiceOptions{
		"yaman": {Token: token, BaseURL: "https://yaman.test"},
	}, WithRoundTripper(rt))
	got, err := rv.Yaman().Secrets().Get("panel-key")
	if err != nil || got["value"] != "secret-value" {
		t.Fatalf("get: %v %v", got, err)
	}
	if calls[0].url != "https://yaman.test/api/v1/secrets/panel-key" {
		t.Fatalf("get url %s", calls[0].url)
	}
	if calls[0].headers["Authorization"] != "Bearer "+token {
		t.Fatal("jwt bearer missing")
	}
	if _, ok := calls[0].headers["X-Api-Key"]; ok {
		t.Fatal("jwt must omit X-Api-Key")
	}
	created, err := rv.Yaman().Secrets().Create(loadFixture(t, "yaman/create-request.json"))
	if err != nil || created["permission"] != "manage" {
		t.Fatalf("create: %v %v", created, err)
	}
	if _, ok := created["value"]; ok {
		t.Fatal("create must omit value")
	}
	if calls[1].url != "https://yaman.test/api/v1/secrets" {
		t.Fatalf("create url %s", calls[1].url)
	}
}

func TestErrorsAndOversize(t *testing.T) {
	var calls []queuedCall
	rt := queuedTrip(t, []string{loadRaw(t, "errors/401.json"), loadRaw(t, "errors/400.json")}, []int{401, 400}, &calls)
	rv := Configure(map[string]ServiceOptions{
		"tracker": {Token: "trk_bad", BaseURL: "https://tracker.test"},
	}, WithRoundTripper(rt))
	_, err := rv.Tracker().Events().List(nil)
	var ae *AuthError
	if !errors.As(err, &ae) || ae.ErrorBody != "Invalid API token." {
		t.Fatalf("auth %v", err)
	}
	_, err = rv.Tracker().Events().Create(map[string]any{})
	var ce *ClientError
	if !errors.As(err, &ce) || ce.ErrorBody != "type is required." {
		t.Fatalf("client %v", err)
	}
	_, err = rv.Tracker().Events().Create(map[string]any{"type": "x", "payload": map[string]any{"blob": strings.Repeat("a", 70000)}})
	var te *TransportError
	if !errors.As(err, &te) {
		t.Fatalf("transport %v", err)
	}
}

func TestSystemsMutations(t *testing.T) {
	var calls []queuedCall
	rt := queuedTrip(t, []string{
		loadRaw(t, "tracker/system-get-response.json"),
		`{}`,
		loadRaw(t, "tracker/system-get-response.json"),
	}, []int{200, 200, 200}, &calls)
	rv := Configure(map[string]ServiceOptions{
		"tracker": {Token: "trk_abc", BaseURL: "https://tracker.test"},
	}, WithRoundTripper(rt))
	if _, err := rv.Tracker().Systems().Update("parent-sys", map[string]any{"name": "Parent Sys"}); err != nil {
		t.Fatal(err)
	}
	if _, err := rv.Tracker().Systems().Delete("parent-sys"); err != nil {
		t.Fatal(err)
	}
	if _, err := rv.Tracker().Systems().SetVersion("parent-sys", "2.0.0"); err != nil {
		t.Fatal(err)
	}
	if calls[0].method != "PUT" || calls[1].method != "DELETE" {
		t.Fatalf("methods %s %s", calls[0].method, calls[1].method)
	}
}

func TestCatalogAndTypedClients(t *testing.T) {
	hosts := loadFixture(t, "catalog/hosts.json")
	var calls []queuedCall
	rt := queuedTrip(t, []string{
		loadRaw(t, "todo/projects-list.json"),
		loadRaw(t, "yaman/list-ok.json"),
		`{}`,
		`{}`,
		loadRaw(t, "policy/check-ok.json"),
		loadRaw(t, "hanap/search-ok.json"),
		loadRaw(t, "ai-center/health.json"),
		`{"ok":true}`,
	}, []int{200, 200, 200, 200, 200, 200, 200, 200}, &calls)
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyIn0.sig"
	rv := Configure(map[string]ServiceOptions{
		"todo":            {Token: "todo_abc"},
		"yaman":           {Token: jwt},
		"policy-manager":  {Token: "pol_abc"},
		"hanap":           {Token: "han_abc"},
	}, WithRoundTripper(rt))
	projects, err := rv.Todo().Projects().List(map[string]string{"status": "active"})
	if err != nil || projects["projects"].([]any)[0].(map[string]any)["slug"] != "desk" {
		t.Fatalf("todo %v %v", projects, err)
	}
	if !strings.HasPrefix(calls[0].url, hosts["todo"].(string)+"/api/v1/projects") {
		t.Fatalf("todo url %s", calls[0].url)
	}
	listed, err := rv.Yaman().Secrets().List()
	if err != nil || listed["secrets"].([]any)[0].(map[string]any)["name"] != "panel-key" {
		t.Fatalf("yaman list %v %v", listed, err)
	}
	if _, err := rv.Yaman().Secrets().Update("panel-key", map[string]any{"description": "x"}); err != nil {
		t.Fatal(err)
	}
	if _, err := rv.Yaman().Secrets().Delete("panel-key"); err != nil {
		t.Fatal(err)
	}
	decision, err := rv.PolicyManager().Check(map[string]any{"action": "queue.submit"})
	if err != nil || decision["allow"] != true {
		t.Fatalf("policy %v %v", decision, err)
	}
	if calls[4].url != hosts["policy-manager"].(string)+"/api/v1/check" {
		t.Fatalf("policy url %s", calls[4].url)
	}
	if _, err := rv.Hanap().Search(map[string]any{"q": "invoice"}); err != nil {
		t.Fatal(err)
	}
	if calls[5].url != "https://hanap.raspberry-vision.cloud/v1/search" {
		t.Fatalf("hanap url %s", calls[5].url)
	}
	if _, err := rv.AiCenter().Health(); err != nil {
		t.Fatal(err)
	}
	if calls[6].url != "https://ai-center.raspberry-vision.cloud"+hosts["ai-center-health"].(string) {
		t.Fatalf("ai-center url %s", calls[6].url)
	}
	if _, err := rv.DatasetBuilder().Health(); err != nil {
		t.Fatal(err)
	}
	if calls[7].url != hosts["dataset-builder"].(string)+"/health" {
		t.Fatalf("dataset url %s", calls[7].url)
	}
	if DefaultBaseURL("sandbox-payments") != hosts["sandbox-payments"].(string) {
		t.Fatal("payments host")
	}
}
