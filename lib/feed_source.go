package murmur

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

type FeedSourceConfig struct {
	Url      string `yaml:"url"`
	Template string `yaml:"template"`
}

type FeedSource struct {
	config *FeedSourceConfig
}

type feedItemSlice []*gofeed.Item

func (a feedItemSlice) Len() int {
	return len(a)
}

func (a feedItemSlice) Less(i, j int) bool {
	if a[i].PublishedParsed != nil && a[j].PublishedParsed != nil {
		return a[i].PublishedParsed.After(*a[i].PublishedParsed)
	}
	if a[i].UpdatedParsed != nil && a[j].UpdatedParsed != nil {
		return a[i].UpdatedParsed.After(*a[i].UpdatedParsed)
	}
	return i < j
}

func (a feedItemSlice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (c *FeedSourceConfig) NewSource() (Source, error) {
	return &FeedSource{
		config: c,
	}, nil
}

func (s *FeedSource) Items(recentUrls []string) ([]*Item, error) {
	resp, err := http.Get(s.config.Url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fp := gofeed.NewParser()
	feed, err := fp.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	return s.itemsFromFeed(feed, recentUrls)
}

func (s *FeedSource) itemsFromFeed(feed *gofeed.Feed, recentUrls []string) ([]*Item, error) {
	t := time.Now().AddDate(0, 0, -1)

	items := feedItemSlice(feed.Items)
	sort.Sort(sort.Reverse(items))

	duplications := make(map[string]struct{}, len(items)+len(recentUrls))
	for _, url := range recentUrls {
		duplications[url] = struct{}{}
	}

	result := make([]*Item, 0, len(items))
	for _, item := range items {
		if item.PublishedParsed != nil && !item.PublishedParsed.After(t) {
			continue
		}
		if item.UpdatedParsed != nil && !item.UpdatedParsed.After(t) {
			continue
		}

		if item.Link == "" {
			continue
		}

		if _, ok := duplications[item.Link]; ok {
			continue
		}
		duplications[item.Link] = struct{}{}

		replacer := strings.NewReplacer("{title}", item.Title, "{url}", item.Link)
		summary := replacer.Replace(s.config.Template)
		result = append(result, &Item{
			Summary: summary,
			Url:     item.Link,
		})
	}
	return result, nil
}
