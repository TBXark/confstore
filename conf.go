package confstore

func Default() Provider {
	return NewConfigProviderGroup(
		NewHttpConfigProvider(JsonCodec{}),
		NewLocalConfigProvider(JsonCodec{}),
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
