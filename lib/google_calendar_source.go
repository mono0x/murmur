package murmur

import (
	"context"
	"io/ioutil"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
	"mvdan.cc/xurls/v2"
)

type GoogleCalendarSourceConfig struct {
	CalendarId   string `yaml:"calendar_id"`
	Template     string `yaml:"template"`
	EndedMessage string `yaml:"ended_message"`
	TimeZone     string `yaml:"time_zone"`
}

type GoogleCalendarSource struct {
	config *GoogleCalendarSourceConfig
	now    time.Time
}

func (c *GoogleCalendarSourceConfig) NewSource() (Source, error) {
	return &GoogleCalendarSource{
		config: c,
		now:    time.Now(),
	}, nil
}

func (s *GoogleCalendarSource) Items() ([]*Item, error) {
	json, err := ioutil.ReadFile("google_client_credentials.json")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	config, err := google.JWTConfigFromJSON(json, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	client := config.Client(context.Background())

	service, err := calendar.New(client)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	updatedMin := s.now.AddDate(0, 0, -1).Format(time.RFC3339)

	events, err := service.Events.List(s.config.CalendarId).UpdatedMin(updatedMin).MaxResults(100).SingleEvents(true).Do()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return s.itemsFromEvents(events)
}

func (s *GoogleCalendarSource) itemsFromEvents(events *calendar.Events) ([]*Item, error) {
	loc, err := time.LoadLocation(events.TimeZone)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var locIn *time.Location
	if s.config.TimeZone != "" {
		var err error
		locIn, err = time.LoadLocation(s.config.TimeZone)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	urlRe := xurls.Strict()
	ignoreBorder := s.now.AddDate(0, 0, -7) // ignore events ended before this time
	items := make([]*Item, 0, len(events.Items))
	for _, event := range events.Items {
		if event.Visibility == "private" {
			continue
		}
		if event.Status == "cancelled" {
			continue
		}

		var timeZone string
		if s.config.TimeZone != "" {
			timeZone = s.config.TimeZone
		} else if event.Start.TimeZone != "" {
			timeZone = event.Start.TimeZone
		} else {
			timeZone = events.TimeZone
		}

		link := event.HtmlLink
		if timeZone != "" {
			u, err := url.Parse(link)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			query := u.Query()
			query.Set("ctz", timeZone)
			u.RawQuery = query.Encode()
			link = u.String()
		}

		var startLoc *time.Location
		if event.Start.TimeZone != "" {
			startLoc, err = time.LoadLocation(event.Start.TimeZone)
			if err != nil {
				return nil, errors.WithStack(err)
			}
		} else {
			startLoc = loc
		}

		var endLoc *time.Location
		if event.End.TimeZone != "" {
			endLoc, err = time.LoadLocation(event.End.TimeZone)
			if err != nil {
				return nil, errors.WithStack(err)
			}
		} else {
			endLoc = loc
		}

		var (
			date    string
			isEnded bool
		)
		if event.Start.Date != "" {
			if !event.EndTimeUnspecified {
				end, err := time.ParseInLocation("2006-01-02", event.End.Date, endLoc)
				if err != nil {
					return nil, errors.WithStack(err)
				}
				if ignoreBorder.After(end) {
					continue
				}
				isEnded = s.now.After(end)
			}

			start, err := time.ParseInLocation("2006-01-02", event.Start.Date, startLoc)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			if locIn != nil {
				start = start.In(locIn)
			}

			date = start.Format("01/02")
		} else if event.Start.DateTime != "" {
			if !event.EndTimeUnspecified {
				end, err := time.ParseInLocation(time.RFC3339, event.End.DateTime, endLoc)
				if err != nil {
					return nil, errors.WithStack(err)
				}
				if ignoreBorder.After(end) {
					continue
				}
				isEnded = s.now.After(end)
			}

			start, err := time.ParseInLocation(time.RFC3339, event.Start.DateTime, startLoc)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			if locIn != nil {
				start = start.In(locIn)
			}

			date = start.Format("01/02")
		}

		var endedMessage string
		if isEnded {
			endedMessage = s.config.EndedMessage
		}

		urls := make([]string, 0)
		if url := urlRe.FindString(event.Description); url != "" {
			urls = append(urls, url)
		}
		urls = append(urls, link)

		location := strings.SplitN(event.Location, ",", 2)[0]
		replacer := strings.NewReplacer(
			"{title}", event.Summary,
			"{url}", link,
			"{urls}", strings.Join(urls, " "),
			"{date}", date,
			"{location}", location,
			"{ended_message}", endedMessage,
		)
		summary := replacer.Replace(s.config.Template)
		items = append(items, &Item{
			Summary: summary,
			Url:     link,
		})
	}

	return items, nil
}
