//nolint:dupl
package plugin

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/ebogdanov/emu-oncall/internal/user"
)

type TextOutput struct {
	logger *zerolog.Logger
}

func NewTextOutput(l *zerolog.Logger) *TextOutput {
	return &TextOutput{logger: l}
}

func (t *TextOutput) CallPhone(_ context.Context, u user.Item, alertID, alertText string) error {
	t.logger.Debug().Msgf("Phone call to %s with text %s, alert id: %s", u.PhoneNumber, alertText, alertID)
	return nil
}

func (t *TextOutput) SendSms(_ context.Context, u user.Item, alertID, alertText string) error {
	t.logger.Debug().Msgf("Send sms to %s with text %s, alert id: %s", u.PhoneNumber, alertText, alertID)
	return nil
}

func (t *TextOutput) MessageSlack(_ context.Context, u user.Item, alertID, alertText string) error {
	t.logger.Debug().Msgf("Send slack message to dbUser with id %s with text %s, alert id: %s", u.ID, alertText, alertID)
	return nil
}
