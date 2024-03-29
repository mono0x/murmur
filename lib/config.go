package murmur

import (
	"fmt"
	"io"

	yaml "gopkg.in/yaml.v2"
)

type SourceType string

const (
	GoogleCalendarSourceType SourceType = "google_calendar"
	FeedSourceType           SourceType = "feed"
)

var SourceConfigs = map[SourceType]func(*Config) SourceConfig{
	GoogleCalendarSourceType: func(c *Config) SourceConfig { return c.Source.GoogleCalendar },
	FeedSourceType:           func(c *Config) SourceConfig { return c.Source.Feed },
}

type SinkType string

const (
	TwitterSinkType SinkType = "twitter"
)

var SinkConfigs = map[SinkType]func(*Config) SinkConfig{
	TwitterSinkType: func(c *Config) SinkConfig { return c.Sink.Twitter },
}

type Config struct {
	Source struct {
		Type           SourceType                  `yaml:"type"`
		GoogleCalendar *GoogleCalendarSourceConfig `yaml:"google_calendar"`
		Feed           *FeedSourceConfig           `yaml:"feed"`
	} `yaml:"source"`
	Sink struct {
		Type    SinkType           `yaml:"type"`
		Twitter *TwitterSinkConfig `yaml:"twitter"`
	} `yaml:"sink"`
}

func LoadConfig(reader io.Reader) (*Config, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return &config, nil
}

func LoadBulkConfig(reader io.Reader) ([]*Config, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	var configs []*Config
	if err := yaml.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return configs, nil
}

func (c *Config) NewSource() (Source, error) {
	if f, ok := SourceConfigs[c.Source.Type]; ok {
		return f(c).NewSource()
	} else {
		return nil, fmt.Errorf("invalid source type: %v", string(c.Source.Type))
	}
}

func (c *Config) NewSink() (Sink, error) {
	if f, ok := SinkConfigs[c.Sink.Type]; ok {
		return f(c).NewSink()
	} else {
		return nil, fmt.Errorf("invalid sink type: %v", string(c.Sink.Type))
	}
}
