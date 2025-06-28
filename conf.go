package confstore

import "net/http"

type options struct {
	httpClient *http.Client
}

// Option is a functional option for configuring default providers in Load/Save.
type Option func(*options)

// WithHTTPClientOption sets a custom http.Client for remote HTTP access.
func WithHTTPClientOption(client *http.Client) Option {
	return func(o *options) {
		o.httpClient = client
	}
}

// defaultProvider returns the default ProviderGroup according to Option
func defaultProvider() Provider {
	return NewProviderGroup(
		NewHttpProvider(JsonCodec{}),
		NewLocalProvider(JsonCodec{}),
	)
}

func defaultProviderWithOpts(optList ...Option) Provider {
	if len(optList) == 0 {
		return defaultProvider()
	}
	o := &options{}
	for _, opt := range optList {
		opt(o)
	}
	return NewProviderGroup(
		NewHttpProvider(JsonCodec{},
			func(p *HttpProvider) {
				if o.httpClient != nil {
					WithHTTPClient(o.httpClient)(p)
				}
			},
		),
		NewLocalProvider(JsonCodec{}),
	)
}

func New(codec Codec) Provider {
	return NewProviderGroup(
		NewHttpProvider(codec),
		NewLocalProvider(codec),
	)
}

/*
Load loads config data from the given path using the default provider group.
Optional functional options can be passed to customize remote access, such as setting the HTTP client.

Example:

	cfg, err := Load[MyType]("http://...", WithHTTPClientOption(myClient))
*/
func Load[T any](path string, opts ...Option) (*T, error) {
	config := new(T)
	err := defaultProviderWithOpts(opts...).Load(path, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

/*
Save saves config data to the given path using the default provider group.
Optional functional options can be passed to customize remote access, such as setting the HTTP client.

Example:

	err := Save("http://...", conf, WithHTTPClientOption(myClient))
*/
func Save[T any](path string, conf *T, opts ...Option) error {
	return defaultProviderWithOpts(opts...).Save(path, conf)
}
