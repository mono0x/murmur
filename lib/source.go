package murmur

type Source interface {
	Items(recentUrls []string) ([]*Item, error)
}

type SourceConfig interface {
	NewSource() (Source, error)
}
