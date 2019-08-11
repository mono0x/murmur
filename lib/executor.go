package murmur

func Execute(config *Config) error {
	source, err := config.NewSource()
	if err != nil {
		return err
	}

	sink, err := config.NewSink()
	if err != nil {
		return err
	}
	defer sink.Close()

	notifier := NewNotifier(source, sink)
	return notifier.Notify()
}
