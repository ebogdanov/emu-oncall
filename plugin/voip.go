//nolint:dupl
package plugin

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/ebogdanov/emu-oncall/internal/user"
)

type Voip struct {
	logger *zerolog.Logger
}

func NewVoip(l *zerolog.Logger) *Voip {
	return &Voip{logger: l}
}

func (t *Voip) CallPhone(_ context.Context, u user.Item, alertID, alertText string) error {
	t.logger.Debug().Msgf("Phone call to %s with text %s, alert id: %s", u.PhoneNumber, alertText, alertID)
	return nil
}

func (t *Voip) SendSms(_ context.Context, u user.Item, alertID, alertText string) error {
	t.logger.Debug().Msgf("Send sms to %s with text %s, alert id: %s", u.PhoneNumber, alertText, alertID)
	return nil
}

func (t *Voip) MessageSlack(_ context.Context, u user.Item, alertID, alertText string) error {
	t.logger.Debug().Msgf("Send slack message to dbUser with id %s with text %s, alert id: %s", u.ID, alertText, alertID)
	return nil
}
