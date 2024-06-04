package grafana

import (
	"bytes"
	"context"
	"time"

	"github.com/ebogdanov/emu-oncall/internal/metrics"

	"github.com/luxifer/ical"
)

type Ics struct {
	iCalURL string
	pm      *metrics.Storage
}

func NewIcsInstance(iCalURL string, pm *metrics.Storage) *Ics {
	client := &Ics{
		iCalURL: iCalURL,
		pm:      pm,
	}
	return client
}

// Current returns who on call now
func (i *Ics) Current(ctx context.Context) (*ScheduleItem, error) {
	httpClient := newAPIClient("", i.pm)

	// Fetch the iCal file from the URL
	response, err := httpClient.Get(ctx, i.iCalURL)
	if err != nil {
		return nil, err
	}

	body := bytes.NewReader(response)
	calendar, err := ical.Parse(body, nil)
	if err != nil {
		return nil, err
	}

	name := ""
	timeZone := "UTC"
	for _, prop := range calendar.Properties {
		switch prop.Name {
		case "X-WR-CALNAME":
			name = prop.Value
		case "X-WR-TIMEZONE":
			timeZone = prop.Value
		}

		if timeZone != "" && name != "" {
			break
		}
	}

	var curr *ScheduleItem

	nowUnix := time.Now().Unix()
	for _, item := range calendar.Events {
		// Filter out old events
		if item.EndDate.Unix() <= nowUnix {
			continue
		}

		if nowUnix < item.StartDate.Unix() {
			break
		}

		if curr == nil {
			curr = &ScheduleItem{
				ID:       item.UID,
				TeamID:   item.Summary,
				TimeZone: timeZone,
				Name:     name,
				Users:    []string{item.Summary},
				Shifts:   []string{item.StartDate.String(), item.EndDate.String()},
			}
		} else {
			curr.Users = append(curr.Users, item.Summary)
		}
	}

	if curr != nil {
		return curr, nil
	}

	return nil, errNotAssigneeSchedule
}
