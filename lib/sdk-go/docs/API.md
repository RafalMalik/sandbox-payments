# API — github.com/raspberry-vision/sdk-go

Go 1.22. Module `github.com/raspberry-vision/sdk-go`. Contract: workspace `libs/docs/platform-sdk.md`. Usage: `libs/docs/platform-sdk-usage.md`.

`Configure(map[string]ServiceOptions, ...Option) *Client`

Typed: `Tracker()`, `Yaman()`, `Todo()`, `Desk()`, `Scheduler()`, `Pulse()`, `Postern()`, `Invest()`, `Abiso()`, `Bisita()`, `Tala()`, `Aklat()`, `Flow()`, `Perfecto()`, `Verdict()`, `Storage()`, `PolicyManager()`, `Hanap()`, `SandboxPayments()`, `Bantay()`, `AiCenter()`, `DatasetBuilder()`, `Service(slug)`.

Exported errors wrap `SDKError`; use `errors.As` for `AuthError`, `ClientError`, `APIError`, `TransportError`.
