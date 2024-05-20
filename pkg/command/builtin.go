package command

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type BuiltIn struct {
	Logger      *zerolog.Logger
	PingHandler func(ctx context.Context) (int, error)
	// PipelineHandler func(ctx context.Context, commands []any) ([]any, error)
}

func (self BuiltIn) Ping(_ struct{}, reply *int) error {
	ctx := self.Logger.WithContext(context.Background())
	ctx = log.Ctx(ctx).With().
		Str("command", "BuiltIn.Ping").
		Logger().WithContext(context.Background())

	var err error
	*reply, err = self.PingHandler(ctx)
	return err
}

// func (self BuiltIn) Pipeline(args []any, reply *[]any) error {
// 	ctx := context.Background()
// 	// ctx := log.Logger.With().
// 	// 	Str("command", "BuiltIn.Ping").
// 	// 	Logger().WithContext(context.Background())

// 	var err error
// 	*reply, err = self.PipelineHandler(ctx, args)
// 	return err
// }

type PingRequest = Request

func NewPingRequest() PingRequest {
	var zero int
	return PingRequest{
		Name:  "BuiltIn.Ping",
		Args:  struct{}{},
		Reply: &zero,
	}
}

type PipelineRequest = Request

func NewPipelineRequest() PipelineRequest {
	return PipelineRequest{
		Name:  "BuiltIn.Pipeline",
		Args:  []any{},
		Reply: &[]any{},
	}
}
