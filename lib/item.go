package murmur

type Location struct {
	Latitude  float64
	Longitude float64
}

type Item struct {
	Summary  string
	Url      string
	Location *Location
}
