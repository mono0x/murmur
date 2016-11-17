package murmur

type Source interface {
	Items() ([]*Item, error)
}

type SourceConfig interface {
	NewSource() (Source, error)
}
