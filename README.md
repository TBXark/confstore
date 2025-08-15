# confstore

Tiny, generic configuration load/save library for Go. Works with local files and HTTP(S), with a simple provider/codec interface so you can extend it as needed.

## Features
- Local and remote (HTTP/S) sources
- JSON codec by default, pluggable via `Codec`
- Composable providers and small API surface

## Install
`go get github.com/TBXark/confstore@latest`

## Quick Start
```go
import (
    "net/http"
    "time"
    confstore "github.com/TBXark/confstore"
)

type App struct {
    Name string `json:"name"`
}

// Load from file or file://
cfg, err := confstore.Load[App]("config.json")

// Load from HTTPS with a custom http.Client
client := &http.Client{Timeout: 5 * time.Second}
remote, err := confstore.Load[App](
    "https://example.com/config.json",
    confstore.WithHTTPClientOption(client),
)

// Save to file or HTTP endpoint
_ = confstore.Save("file://out.json", &remote)
_ = confstore.Save("https://example.com/config", &cfg)
```

## Providers & Codecs
- Providers: `LocalProvider`, `HttpProvider`, or build your own by implementing `Provider`.
- Codecs: built-in `JsonCodec`; add others by implementing `Codec`.
- `ProviderGroup` composes multiple providers; helpers in `conf.go` wire sensible defaults.

## Notes
- `HttpProvider` checks non-2xx responses and sets `Content-Type: application/json` on save.
- For security and reliability, prefer passing your own `http.Client` with timeouts.

## License
**confstore** is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.