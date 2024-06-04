package grafana

import (
	"os"
	"time"

	"github.com/luxifer/ical"
)

func Parse(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	calendar, err := ical.Parse(file, nil)
	if err != nil {
		return err
	}

	for _, item := range calendar.Events {
		if time.Now().Unix() <= item.StartDate.Unix() || item.EndDate.Unix() <= time.Now().Unix() {
			continue
		}
	}

	return nil
}
