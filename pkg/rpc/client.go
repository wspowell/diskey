package rpc

import (
	"context"
	"net"
	"net/rpc"
	"time"

	"github.com/rs/zerolog/log"

	"diskey/pkg/command"
	"diskey/pkg/errors"
)

const (
	defaultSendTimeout     = 3 * time.Second
	defaultReceiveTimeout  = 3 * time.Second
	defaultKeepAlivePeriod = 10 * time.Second
)

type Client struct {
	tcpConnection  *net.TCPConn
	rpcClient      *rpc.Client
	cancel         context.CancelFunc
	host           string
	port           string
	address        string
	sendTimeout    time.Duration
	receiveTimeout time.Duration
}

func NewClient(host string, port string) *Client {
	return &Client{
		host:           host,
		port:           port,
		address:        host + ":" + port,
		tcpConnection:  nil,
		cancel:         nil,
		sendTimeout:    defaultSendTimeout,
		receiveTimeout: defaultReceiveTimeout,
	}
}

func NewWithConnection(tcpConnection *net.TCPConn) *Client {
	return &Client{
		address:        tcpConnection.RemoteAddr().String(),
		tcpConnection:  tcpConnection,
		sendTimeout:    defaultSendTimeout,
		receiveTimeout: defaultReceiveTimeout,
	}
}

func (self *Client) Disconnect(ctx context.Context) {
	if self.cancel != nil {
		self.cancel()
		self.cancel = nil
	}

	if self.tcpConnection != nil {
		if err := self.tcpConnection.Close(); err != nil {
			log.Ctx(ctx).Err(err).Msg("failed to close connection")
		}
		log.Ctx(ctx).Debug().Msg("closed client connection")
		self.tcpConnection = nil
		self.rpcClient = nil
	}
}

func (self *Client) Host() string {
	return self.host
}

func (self *Client) Port() string {
	return self.port
}

func (self *Client) Address() string {
	return self.address
}

func (self *Client) SetSendTimeout(sendTimeout time.Duration) {
	self.sendTimeout = sendTimeout
}

func (self *Client) SetReceiveTimeout(receiveTimeout time.Duration) {
	self.receiveTimeout = receiveTimeout
}

func (self *Client) Connect(ctx context.Context) errors.Error[ConnectError] {
	ctx, cancel := context.WithCancel(ctx)
	self.cancel = cancel

	tcpAddr, err := net.ResolveTCPAddr("tcp", self.address)
	if err != nil {
		return errors.NewWithErr(ConnectErrorInvalidAddress, err)
	}

	tcpConnection, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return errors.NewWithErr(ConnectErrorConnectionFailure, err)
	}

	ctx = log.Ctx(ctx).With().
		Str("source", "client").
		Str("remoteAddress", tcpConnection.RemoteAddr().String()).
		Logger().WithContext(ctx)

	self.tcpConnection = tcpConnection
	if keepAliveErr := self.tcpConnection.SetKeepAlive(true); keepAliveErr != nil {
		return errors.NewWithErr(ConnectErrorConnectionFailure, keepAliveErr)
	}
	if keepAliveErr := self.tcpConnection.SetKeepAlivePeriod(defaultKeepAlivePeriod); keepAliveErr != nil {
		return errors.NewWithErr(ConnectErrorConnectionFailure, keepAliveErr)
	}
	if keepAliveErr := self.tcpConnection.SetNoDelay(false); keepAliveErr != nil {
		return errors.NewWithErr(ConnectErrorConnectionFailure, keepAliveErr)
	}
	log.Ctx(ctx).Debug().Msg("connected")

	self.rpcClient = rpc.NewClientWithCodec(NewClientCodecMsgpack(self.tcpConnection))

	go func(ctx context.Context) {
		<-ctx.Done()
		self.Disconnect(ctx)
	}(ctx)

	return errors.Ok[ConnectError]()
}

func (self *Client) Send(ctx context.Context, cmd command.Request) error {
	if self.tcpConnection == nil {
		log.Ctx(ctx).Error().Msg("not connected")
		return nil
	}

	if self.rpcClient == nil {
		log.Ctx(ctx).Error().Msg("not connected")
		return nil
	}

	if err := self.tcpConnection.SetDeadline(time.Now().Add(self.receiveTimeout)); err != nil {
		log.Ctx(ctx).Err(err).Msg("failed to set deadline")
		return nil
	}

	return command.Send(ctx, self.rpcClient, cmd)
}
