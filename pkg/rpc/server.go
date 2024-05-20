package rpc

import (
	"context"
	"net"
	"net/rpc"
	"time"

	"github.com/rs/zerolog/log"

	"diskey/pkg/errors"
)

type Server struct {
	rpcServer      *rpc.Server
	host           string
	port           string
	address        string
	keepAlive      time.Duration
	sendTimeout    time.Duration
	receiveTimeout time.Duration
}

func NewServer(host string, port string) Server {
	return Server{
		rpcServer:      rpc.NewServer(),
		host:           host,
		port:           port,
		address:        host + ":" + port,
		keepAlive:      defaultKeepAlivePeriod,
		sendTimeout:    defaultSendTimeout,
		receiveTimeout: defaultReceiveTimeout,
	}
}

func (self Server) Address() string {
	return self.address
}

func (self Server) Host() string {
	return self.host
}

func (self Server) Port() string {
	return self.port
}

// RegisterHandler publishes in the server the set of methods of the
// receiver value that satisfy the following conditions:
//   - exported method of exported type
//   - two arguments, both of exported type
//   - the second argument is a pointer
//   - one return value, of type error
//
// It returns an error if the receiver is not an exported type or has
// no suitable methods. It also logs the error using package log.
// The client accesses each method using a string of the form "Type.Method",
// where Type is the receiver's concrete type.
func (self Server) RegisterHandler(ctx context.Context, handler any) errors.Error[HandlersError] {
	if rpcErr := self.rpcServer.Register(handler); rpcErr != nil {
		log.Ctx(ctx).Err(rpcErr).Type("handlerType", handler).Msg("failed to register rpc extension handler")
		return errors.NewWithErr(HandlersErrorRpcFailure, rpcErr)
	}

	return errors.Ok[HandlersError]()
}

func (self Server) Listen(ctx context.Context) (net.Listener, errors.Error[ListenError]) {
	ctx = log.Ctx(ctx).With().
		Str("source", "server").
		Str("serverHost", self.host).
		Str("serverPort", self.port).
		Logger().WithContext(ctx)

	if rpcErr := self.rpcServer.Register(newBuiltInHandlers(ctx).BuiltIn); rpcErr != nil {
		log.Ctx(ctx).Err(rpcErr).Msg("failed to register rpc built in handlers")
		return nil, errors.NewWithErr(ListenErrorRpcFailure, rpcErr)
	}

	listenConfig := &net.ListenConfig{
		Control:   nil,
		KeepAlive: self.keepAlive,
	}

	netListener, err := listenConfig.Listen(ctx, "tcp", self.address)
	if err != nil {
		return nil, errors.NewWithErr(ListenErrorListenFailure, err)
	}

	log.Ctx(ctx).Debug().Msg("listening for connections")

	return netListener, errors.Ok[ListenError]()
}

func (self Server) AcceptConnections(ctx context.Context, netListener net.Listener) {
	ctx = log.Ctx(ctx).With().
		Str("source", "server").
		Str("serverHost", self.host).
		Str("serverPort", self.port).
		Logger().WithContext(ctx)

	go func() {
		<-ctx.Done()
		_ = netListener.Close()
	}()

	var connectionId uint64
	var netConnection net.Conn
	var err error

	log.Ctx(ctx).Debug().Msg("accepting new connections")
	for {
		if ctx.Err() != nil {
			break
		}

		netConnection, err = netListener.Accept()
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				break
			}

			// FIXME: This gets hit with "use of closed network connection"
			// log.Ctx(ctx).Err(err).Msg("failed to accept connection")
			continue
		}

		asTcpConnection, ok := netConnection.(*net.TCPConn)
		if !ok {
			log.Ctx(ctx).Error().Msg("connection must be tcp connection")
			continue
		}

		connectionId++

		connectionCtx := log.Ctx(ctx).With().
			Uint64("connectionId", connectionId).
			Str("clientAddress", asTcpConnection.RemoteAddr().String()).
			Logger().WithContext(ctx)

		if keepAliveErr := asTcpConnection.SetKeepAlive(true); keepAliveErr != nil {
			// return errors.NewWithErr(ConnectErrorConnectionFailure, keepAliveErr)
			continue
		}
		if keepAliveErr := asTcpConnection.SetKeepAlivePeriod(defaultKeepAlivePeriod); keepAliveErr != nil {
			// return errors.NewWithErr(ConnectErrorConnectionFailure, keepAliveErr)
			continue
		}
		if keepAliveErr := asTcpConnection.SetNoDelay(false); keepAliveErr != nil {
			// return errors.NewWithErr(ConnectErrorConnectionFailure, keepAliveErr)
			continue
		}

		go handleConnection(connectionCtx, asTcpConnection, self.rpcServer)
	}
}

func handleConnection(_ context.Context, tcpConnection *net.TCPConn, rpcServer *rpc.Server) {
	// log.Ctx(ctx).Debug().Msg("handling new connection")

	// go func(ctx context.Context, tcpConnection *net.TCPConn) {
	// 	<-ctx.Done()
	// 	_ = tcpConnection.Close()
	// }(ctx, tcpConnection)

	rpcServer.ServeCodec(NewServerCodecMsgpack(tcpConnection))

	// log.Ctx(ctx).Debug().Msg("closed server connection")
}
