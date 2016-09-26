package murmur

type Sink interface {
	Close()
	RecentUrls() ([]string, error)
	Output(item *Item) error
}

type SinkConfig interface {
	NewSink() (Sink, error)
}
