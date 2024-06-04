package events

import (
	"context"
	"time"

	"github.com/ebogdanov/emu-oncall/internal/db"
	"github.com/rs/zerolog"
)

const (
	waitTime = 100 * time.Millisecond
)

type Service interface {
	Add(i *Record) error
}

type Record struct {
	Timestamp time.Time
	UserID    string
	Recipient string
	Channel   string
	Success   bool
	Msg       string
}

type DefaultService struct {
	pool   chan *Record
	stop   chan bool
	db     *db.DBx
	logger zerolog.Logger
}

func New(dbShard *db.DBx, l zerolog.Logger) *DefaultService {
	return &DefaultService{
		pool:   make(chan *Record, 10000),
		stop:   make(chan bool),
		db:     dbShard,
		logger: l.With().Str("component", "events").Logger(),
	}
}

func (d *DefaultService) Listen(ctx context.Context) {
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
}

func (d *DefaultService) Stop() {
	d.stop <- true
}

func (d *DefaultService) Add(item *Record) error {
	d.pool <- item

	return nil
}

func (d *DefaultService) insert(item *Record) {
	msg := item.Msg
	if len(msg) > 500 {
		msg = msg[:500]
	}

	_, err := d.db.ExecContext(context.Background(),
		"INSERT INTO events (date_add, user_id, channel, recipient, success, msg) VALUES ($1, $2, $3, $4, $5, $6)",
		&item.Timestamp, &item.UserID, &item.Channel, &item.Recipient, &item.Success, &msg)

	if err != nil {
		d.logger.Error().Err(err).Msg("unable to insert events into database")
	}
}
