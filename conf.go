package confstore

func Default() Provider {
	return NewProviderGroup(
		NewHttpProvider(JsonCodec{}),
		NewLocalProvider(JsonCodec{}),
	)
}

func New(codec Codec) Provider {
	return NewProviderGroup(
		NewHttpProvider(codec),
		NewLocalProvider(codec),
	)
}

func Load[T any](path string) (*T, error) {
	config := new(T)
	err := Default().Load(path, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func Save[T any](path string, conf *T) error {
	return Default().Save(path, conf)
}
