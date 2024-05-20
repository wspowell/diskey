package command

import (
	"context"
	"fmt"
	"net/rpc"
)

type Request struct {
	Args  any
	Reply any
	Name  string
}

func Send(_ context.Context, rpcClient *rpc.Client, request Request) error {
	if request.Name == "" {
		return fmt.Errorf("request name cannot be blank")
	}

	if request.Args == nil {
		return fmt.Errorf("request args cannot be nil")
	}

	if request.Reply == nil {
		return fmt.Errorf("request reply cannot be nil")
	}

	return rpcClient.Call(request.Name, request.Args, request.Reply)
}
