package murmur

import (
	"context"
	"io/ioutil"
	"strings"
	"time"

	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
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
		return nil, err
	}

	config, err := google.JWTConfigFromJSON(json, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, err
	}

	client := config.Client(context.Background())

	service, err := calendar.New(client)
	if err != nil {
		return nil, err
	}

	updatedMin := s.now.AddDate(0, 0, -1).Format(time.RFC3339)

	events, err := service.Events.List(s.config.CalendarId).UpdatedMin(updatedMin).MaxResults(100).SingleEvents(true).Do()
	if err != nil {
		return nil, err
	}

	return s.itemsFromEvents(events)
}

func (s *GoogleCalendarSource) itemsFromEvents(events *calendar.Events) ([]*Item, error) {
	ignoreBorder := s.now.AddDate(0, 0, -7) // ignore events ended before this time
	items := make([]*Item, 0, len(events.Items))
	for _, event := range events.Items {
		if event.Status == "cancelled" {
			continue
		}
		link := event.HtmlLink
		if s.config.TimeZone != "" {
			link += "&ctz=" + s.config.TimeZone
		}

		var (
			date    string
			isEnded bool
		)
		if event.Start.Date != "" {
			endLoc, err := time.LoadLocation(event.End.TimeZone)
			if err != nil {
				return nil, err
			}
			end, err := time.ParseInLocation("2006-01-02", event.End.Date, endLoc)
			if err != nil {
				return nil, err
			}
			if ignoreBorder.After(end) {
				continue
			}

			startLoc, err := time.LoadLocation(event.Start.TimeZone)
			if err != nil {
				return nil, err
			}
			start, err := time.ParseInLocation("2006-01-02", event.Start.Date, startLoc)
			if err != nil {
				return nil, err
			}

			date = start.Format("01/02")
			isEnded = s.now.After(end)
		} else if event.Start.DateTime != "" {
			endLoc, err := time.LoadLocation(event.End.TimeZone)
			if err != nil {
				return nil, err
			}
			end, err := time.ParseInLocation(time.RFC3339, event.End.DateTime, endLoc)
			if err != nil {
				return nil, err
			}
			if ignoreBorder.After(end) {
				continue
			}

			startLoc, err := time.LoadLocation(event.Start.TimeZone)
			if err != nil {
				return nil, err
			}
			start, err := time.ParseInLocation(time.RFC3339, event.Start.DateTime, startLoc)
			if err != nil {
				return nil, err
			}

			date = start.Format("01/02")
			isEnded = s.now.After(end)
		}

		var endedMessage string
		if isEnded {
			endedMessage = s.config.EndedMessage
		}

		location := strings.SplitN(event.Location, ",", 2)[0]
		replacer := strings.NewReplacer(
			"{title}", event.Summary,
			"{url}", link,
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
