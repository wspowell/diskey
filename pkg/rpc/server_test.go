package rpc_test

import (
	"context"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	"diskey/pkg/command"
	"diskey/pkg/errors/errorstest"
	"diskey/pkg/rpc"
)

func Test_Server_Handler(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	port := "7000"

	testServer := rpc.NewServer("localhost", port)
	handlerErr := testServer.RegisterHandler(ctx, TestExtension{
		extensionHandler: func(_ context.Context, input string) (TestHandlerReply, error) {
			if input == "test" {
				return TestHandlerReply{
					Output: "success",
				}, nil
			}

			return TestHandlerReply{
				Output: "failure",
			}, nil
		},
	})
	errorstest.NoError(t, handlerErr)

	listener, listenErr := testServer.Listen(ctx)
	errorstest.NoError(t, listenErr)

	go testServer.AcceptConnections(ctx, listener)

	testClient := rpc.NewClient("localhost", port)
	connectErr := testClient.Connect(ctx)
	errorstest.NoError(t, connectErr)

	one := 1

	tests := []struct {
		name          string
		request       command.Request
		expectedReply any
	}{
		{
			name:          "built in - ping",
			request:       command.NewPingRequest(),
			expectedReply: &one,
		},
		{
			name:    "extension",
			request: newExtensionRequest("test"),
			expectedReply: &TestHandlerReply{
				Output: "success",
			},
		},
	}
	for index := range tests {
		testCase := tests[index]
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testClient.Send(ctx, testCase.request)
			assert.NoError(t, err)

			assert.Equal(t, testCase.expectedReply, testCase.request.Reply)
		})
	}
}

type TestExtension struct {
	extensionHandler func(ctx context.Context, input string) (TestHandlerReply, error)
}

type TestHandlerArgs struct {
	Input string
}

type TestHandlerReply struct {
	Output string
}

func (self TestExtension) Extension(args TestHandlerArgs, reply *TestHandlerReply) error {
	ctx := log.Logger.With().
		Str("command", "TestExtension.Extension").
		Logger().WithContext(context.Background())

	var err error
	*reply, err = self.extensionHandler(ctx, args.Input)
	return err
}

type extensionRequest = command.Request

func newExtensionRequest(input string) extensionRequest {
	return extensionRequest{
		Name:  "TestExtension.Extension",
		Args:  TestHandlerArgs{Input: input},
		Reply: &TestHandlerReply{},
	}
}
