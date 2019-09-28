package murmur

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

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
	api := anaconda.NewTwitterApiWithCredentials(
		c.OAuthToken, c.OAuthTokenSecret, c.ConsumerKey, c.ConsumerSecret)
	api.HttpClient = &http.Client{Timeout: 10 * time.Second}
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
		return nil, fmt.Errorf("%w", err)
	}

	v := url.Values{}
	v.Set("user_id", userId)
	v.Set("count", "200") // TODO: read from the config
	timeline, err := s.api.GetUserTimeline(v)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	urls := make([]string, 0, len(timeline)) // heuristic optimization
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
		var apiErr *anaconda.ApiError
		if errors.As(err, &apiErr) {
			for _, err := range apiErr.Decoded.Errors {
				if err.Code == anaconda.TwitterErrorStatusIsADuplicate {
					return nil
				}
			}
			return fmt.Errorf("%w", apiErr)
		} else {
			return fmt.Errorf("%w", err)
		}
	}
	return nil
}
