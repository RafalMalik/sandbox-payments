package sdk

const (
	MaxBodyBytes   = 65536
	TimeoutSeconds = 10
)

var DefaultSlugs = []string{
	"tracker", "todo", "desk", "scheduler", "pulse", "postern",
	"invest", "abiso", "bisita", "tala", "aklat", "flow", "yaman",
	"perfecto", "verdict", "bantay", "hanap", "storage", "policy-manager",
	"sandbox-payments", "ai-center", "dataset-builder",
}

var hostOverrides = map[string]string{
	"policy-manager":   "https://policy.raspberry-vision.cloud",
	"dataset-builder":  "https://dataset.raspberry-vision.cloud",
	"sandbox-payments": "https://payments.raspberry-vision.cloud",
}

var healthPaths = map[string]string{
	"ai-center": "/healthz",
}

func DefaultBaseURL(slug string) string {
	if host, ok := hostOverrides[slug]; ok {
		return host
	}
	return "https://" + slug + ".raspberry-vision.cloud"
}

func HealthPath(slug string) string {
	if path, ok := healthPaths[slug]; ok {
		return path
	}
	return "/health"
}
