package rvsdk

import sdk "github.com/raspberry-vision/sdk-go"

// Configure builds a Platform HTTP client. Tokens may be empty for health checks.
func Configure(services map[string]sdk.ServiceOptions) *sdk.Client {
	return sdk.Configure(services)
}
