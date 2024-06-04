package plugin

import (
	"context"

	"github.com/ebogdanov/emu-oncall/internal/user"
)

type Plugin interface {
	CallPhone(context.Context, user.Item, string, string) error
	SendSms(context.Context, user.Item, string, string) error
	MessageSlack(context.Context, user.Item, string, string) error
}
