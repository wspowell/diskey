package rpc

import (
	"context"

	"diskey/pkg/command"

	"github.com/rs/zerolog/log"
)

type builtInHandlers struct {
	command.BuiltIn
}

func newBuiltInHandlers(ctx context.Context) builtInHandlers {
	return builtInHandlers{
		BuiltIn: command.BuiltIn{
			Logger: log.Ctx(ctx),
			PingHandler: func(ctx context.Context) (int, error) {
				log.Ctx(ctx).Debug().Msg("ping")
				return 1, nil
			},
		},
	}
}
