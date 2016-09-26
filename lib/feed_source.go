package murmur

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	"golang.org/x/tools/blog/atom"
)

type FeedSourceConfig struct {
	Url      string `yaml:"url"`
	Template string `yaml:"template"`
}

type FeedSource struct {
	config *FeedSourceConfig
}

type atomEntry struct {
	entry   *atom.Entry
	hasTime bool
	time    time.Time
}

func newAtomEntry(entry *atom.Entry) *atomEntry {
	var t time.Time
	hasTime := false

	if entry.Published != "" {
		if published, err := time.Parse(time.RFC3339, string(entry.Published)); err == nil {
			t = published
			hasTime = true
		}
	} else if entry.Updated != "" {
		if updated, err := time.Parse(time.RFC3339, string(entry.Updated)); err == nil {
			t = updated
			hasTime = true
		}
	}

	return &atomEntry{
		entry:   entry,
		hasTime: hasTime,
		time:    t,
	}
}

type atomEntrySlice []*atomEntry

func (a atomEntrySlice) Len() int {
	return len(a)
}

func (a atomEntrySlice) Less(i, j int) bool {
	if a[i].hasTime && a[j].hasTime {
		return a[i].time.After(a[j].time)
	} else {
		return i < j
	}
}

func (a atomEntrySlice) Swap(i, j int) {
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var feed atom.Feed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, err
	}
	return s.itemsFromFeed(&feed, recentUrls)
}

func (s *FeedSource) itemsFromFeed(feed *atom.Feed, recentUrls []string) ([]*Item, error) {
	t := time.Now().AddDate(0, 0, -1)

	atomEntries := make(atomEntrySlice, 0, len(feed.Entry))
	for _, entry := range feed.Entry {
		atomEntries = append(atomEntries, newAtomEntry(entry))
	}
	sort.Sort(sort.Reverse(atomEntries))

	duplications := make(map[string]struct{}, len(atomEntries)+len(recentUrls))
	for _, url := range recentUrls {
		duplications[url] = struct{}{}
	}

	items := make([]*Item, 0, len(atomEntries))
	for _, atomEntry := range atomEntries {
		entry := atomEntry.entry

		if atomEntry.hasTime && !atomEntry.time.After(t) {
			continue
		}

		if len(entry.Link) == 0 {
			continue
		}
		link := entry.Link[0]
		if _, ok := duplications[link.Href]; ok {
			continue
		}
		duplications[link.Href] = struct{}{}

		replacer := strings.NewReplacer("{title}", entry.Title, "{url}", link.Href)
		summary := replacer.Replace(s.config.Template)
		items = append(items, &Item{
			Summary: summary,
			Url:     link.Href,
		})
	}
	return items, nil
}
