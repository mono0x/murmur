package murmur

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSourceForFeedSource(t *testing.T) {
	config, err := LoadConfig(strings.NewReader(`---
source:
  type: feed
  feed:
    url: https://example.com/
    template: "{title}: {url}"
`))
	if err != nil {
		t.Fatal(err)
	}

	source, err := config.NewSource()
	if err != nil {
		t.Fatal(err)
	}

	assert.IsType(t, &FeedSource{}, source)
	feedSourceConfig := source.(*FeedSource).config
	assert.Equal(t, "https://example.com/", feedSourceConfig.Url)
	assert.Equal(t, "{title}: {url}", feedSourceConfig.Template)
}

func TestNewSourceForGoogleCalendarSource(t *testing.T) {
	config, err := LoadConfig(strings.NewReader(`---
source:
  type: google_calendar
  google_calendar:
    calendar_id: test
    template: "{title}: {url}"
    time_zone: Asia/Tokyo
`))
	if err != nil {
		t.Fatal(err)
	}

	source, err := config.NewSource()
	if err != nil {
		t.Fatal(err)
	}

	assert.IsType(t, &GoogleCalendarSource{}, source)
	googleCalendarSourceConfig := source.(*GoogleCalendarSource).config
	assert.Equal(t, "test", googleCalendarSourceConfig.CalendarId)
	assert.Equal(t, "{title}: {url}", googleCalendarSourceConfig.Template)
	assert.Equal(t, "Asia/Tokyo", googleCalendarSourceConfig.TimeZone)
}

func TestNewSinkForTwitterSink(t *testing.T) {
	config, err := LoadConfig(strings.NewReader(`---
sink:
  type: twitter
  twitter:
    consumer_key: 01
    consumer_secret: 02
    oauth_token: 03
    oauth_token_secret: 04
`))
	if err != nil {
		t.Fatal(err)
	}

	sink, err := config.NewSink()
	if err != nil {
		t.Fatal(err)
	}

	assert.IsType(t, &TwitterSink{}, sink)
	twitterSinkConfig := sink.(*TwitterSink).config
	assert.Equal(t, "01", twitterSinkConfig.ConsumerKey)
	assert.Equal(t, "02", twitterSinkConfig.ConsumerSecret)
	assert.Equal(t, "03", twitterSinkConfig.OAuthToken)
	assert.Equal(t, "04", twitterSinkConfig.OAuthTokenSecret)
}
