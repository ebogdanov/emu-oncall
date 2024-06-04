package plugin

import (
	"context"

	"github.com/ebogdanov/emu-oncall/internal/logger"
	"github.com/ebogdanov/emu-oncall/internal/user"
)

type RestAPI struct {
	logger *logger.Instance
}

func NewRestAPI(l *logger.Instance) *RestAPI {
	return &RestAPI{logger: l}
}

func (t *RestAPI) CallPhone(_ context.Context, u user.Item, alertID, alertText string) error {
	t.logger.Debug().Msgf("Phone call to %s with text %s, alert id: %s", u.PhoneNumber, alertText, alertID)
	return nil
}

func (t *RestAPI) SendSms(_ context.Context, u user.Item, alertID, alertText string) error {
	t.logger.Debug().Msgf("Send sms to %s with text %s, alert id: %s", u.PhoneNumber, alertText, alertID)
	return nil
}

func (t *RestAPI) MessageSlack(_ context.Context, u user.Item, alertID, alertText string) error {
	t.logger.Debug().Msgf("Send slack message to dbUser with id %s with text %s, alert id: %s", u.ID, alertText, alertID)
	return nil
}
