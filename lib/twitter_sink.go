package murmur

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/ChimeraCoder/anaconda"
)

type TwitterSinkConfig struct {
	ConsumerKey      string `yaml:"consumer_key"`
	ConsumerSecret   string `yaml:"consumer_secret"`
	OAuthToken       string `yaml:"oauth_token"`
	OAuthTokenSecret string `yaml:"oauth_token_secret"`
}

type TwitterSink struct {
	config *TwitterSinkConfig
	api    *anaconda.TwitterApi
}

func (c *TwitterSinkConfig) NewSink() (Sink, error) {
	// XXX: shouldn't update global variables...
	anaconda.SetConsumerKey(c.ConsumerKey)
	anaconda.SetConsumerSecret(c.ConsumerSecret)

	api := anaconda.NewTwitterApi(c.OAuthToken, c.OAuthTokenSecret)
	return &TwitterSink{
		config: c,
		api:    api,
	}, nil
}

func (s *TwitterSink) Close() {
	s.api.Close()
}

func (s *TwitterSink) RecentUrls() ([]string, error) {
	userId := strings.SplitN(s.config.OAuthToken, "-", 2)[0]
	if _, err := strconv.ParseInt(userId, 10, 64); err != nil {
		return nil, err
	}

	v := url.Values{}
	v.Set("user_id", userId)
	v.Set("count", "200") // TODO: read from the config
	timeline, err := s.api.GetUserTimeline(v)
	if err != nil {
		return nil, err
	}

	urls := make([]string, len(timeline))
	for _, status := range timeline {
		for _, url := range status.Entities.Urls {
			urls = append(urls, url.Expanded_url)
		}
	}

	return urls, nil
}

func (s *TwitterSink) Output(item *Item) error {
	values := url.Values{}

	if _, err := s.api.PostTweet(item.Summary, values); err != nil {
		if apiErr, ok := err.(*anaconda.ApiError); ok {
			for _, err := range apiErr.Decoded.Errors {
				if err.Code == anaconda.TwitterErrorStatusIsADuplicate {
					return nil
				}
			}
			return apiErr
		} else {
			return err
		}
	}
	return nil
}
