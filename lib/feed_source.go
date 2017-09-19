package murmur

import (
	"io/ioutil"
	"sort"
	"strings"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/mmcdole/gofeed"
)

type FeedSourceConfig struct {
	Url      string `yaml:"url"`
	Template string `yaml:"template"`
}

type FeedSource struct {
	config *FeedSourceConfig
}

func (c *FeedSourceConfig) NewSource() (Source, error) {
	return &FeedSource{
		config: c,
	}, nil
}

func (s *FeedSource) Items() ([]*Item, error) {
	client := retryablehttp.NewClient()
	client.RetryMax = 1
	client.Logger.SetOutput(ioutil.Discard)

	resp, err := client.Get(s.config.Url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fp := gofeed.NewParser()
	feed, err := fp.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	return s.itemsFromFeed(feed)
}

func (s *FeedSource) itemsFromFeed(feed *gofeed.Feed) ([]*Item, error) {
	t := time.Now().AddDate(0, 0, -1)

	items := make([]*gofeed.Item, 0, len(feed.Items))
	for _, item := range feed.Items {
		if item.PublishedParsed != nil {
			if !item.PublishedParsed.After(t) {
				continue
			}
		} else if item.UpdatedParsed != nil {
			if !item.UpdatedParsed.After(t) {
				continue
			}
		}

		if item.Link == "" {
			continue
		}

		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].PublishedParsed != nil && items[j].PublishedParsed != nil {
			return items[i].PublishedParsed.Before(*items[i].PublishedParsed)
		}
		if items[i].UpdatedParsed != nil && items[j].UpdatedParsed != nil {
			return items[i].UpdatedParsed.Before(*items[i].UpdatedParsed)
		}
		return i > j
	})

	result := make([]*Item, 0, len(items))
	for _, item := range items {
		replacer := strings.NewReplacer("{title}", item.Title, "{url}", item.Link)
		summary := replacer.Replace(s.config.Template)
		result = append(result, &Item{
			Summary: summary,
			Url:     item.Link,
		})
	}
	return result, nil
}
