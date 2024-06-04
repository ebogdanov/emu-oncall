package events

import (
	"context"
	"database/sql"
	"time"

	"github.com/ebogdanov/emu-oncall/internal/logger"
)

const (
	waitTime = 100 * time.Millisecond
)

type Service interface {
	Add(i *Item) error
}

type Item struct {
	Timestamp time.Time
	UserID    string
	Recipient string
	Channel   string
	Success   bool
	Msg       string
}

type DefaultService struct {
	pool   chan *Item
	stop   chan bool
	db     *sql.DB
	logger logger.Instance
}

func New(dbShard *sql.DB, l logger.Instance) *DefaultService {
	return &DefaultService{
		pool:   make(chan *Item, 10000),
		stop:   make(chan bool),
		db:     dbShard,
		logger: l,
	}
}

func (d *DefaultService) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case data := <-d.pool:
				d.insert(data)
			case <-ctx.Done():
				close(d.pool)
				return
			case <-d.stop:
				close(d.pool)
				return
			default:
				time.Sleep(waitTime)
			}
		}
	}()
}

func (d *DefaultService) Stop() {
	d.stop <- true
}

func (d *DefaultService) Add(item *Item) error {
	d.pool <- item

	return nil
}

func (d *DefaultService) insert(item *Item) {
	msg := item.Msg
	if len(msg) > 500 {
		msg = msg[:500]
	}

	_, err := d.db.ExecContext(context.Background(),
		"INSERT INTO events (date_add, user_id, channel, recipient, success, msg) VALUES ($1, $2, $3, $4, $5, $6)",
		&item.Timestamp, &item.UserID, &item.Channel, &item.Recipient, &item.Success, &msg)

	if err != nil {
		d.logger.Error().Str("component", "events").Err(err).Msg("unable to insert events into database")
	}
}
