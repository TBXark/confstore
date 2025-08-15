package confstore

import "net/http"

type options struct {
	httpProviderOptions []HttpProviderOption
}
type Option func(*options)

func WithHTTPClientOption(client *http.Client) Option {
	return func(o *options) {
		o.httpProviderOptions = append(o.httpProviderOptions, WithHTTPClient(client))
	}
}

func newOptions(opts ...Option) *options {
	defaults := &options{
		httpProviderOptions: []HttpProviderOption{},
	}
	for _, opt := range opts {
		opt(defaults)
	}
	return defaults
}

func defaultProvider(options ...Option) Provider {
	opts := newOptions(options...)
	return NewProviderGroup(
		NewHttpProvider(JsonCodec{}, opts.httpProviderOptions...),
		NewLocalProvider(JsonCodec{}),
	)
}

func New(codec Codec, options ...Option) Provider {
	opts := newOptions(options...)
	return NewProviderGroup(
		NewHttpProvider(codec, opts.httpProviderOptions...),
		NewLocalProvider(codec),
	)
}

func Load[T any](path string, opts ...Option) (*T, error) {
	config := new(T)
	provider := defaultProvider(opts...)
	err := provider.Load(path, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func Save[T any](path string, conf *T, opts ...Option) error {
	provider := defaultProvider(opts...)
	return provider.Save(path, conf)
}
