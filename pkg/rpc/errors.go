package rpc

type ListenError uint

const (
	ListenErrorListenFailure = ListenError(iota + 1)
	ListenErrorRpcFailure
)

func (self ListenError) String() string {
	switch self {
	case ListenErrorListenFailure:
		return "ListenFailure"
	case ListenErrorRpcFailure:
		return "RpcFailure"
	default:
		return "ListenError"
	}
}

type HandlersError uint

const (
	HandlersErrorRpcFailure = HandlersError(iota + 1)
)

func (self HandlersError) String() string {
	switch self {
	case HandlersErrorRpcFailure:
		return "RpcFailure"
	default:
		return "HandlersError"
	}
}

type ConnectError uint

const (
	ConnectErrorInvalidAddress = ConnectError(iota + 1)
	ConnectErrorConnectionFailure
)

func (self ConnectError) String() string {
	switch self {
	case ConnectErrorInvalidAddress:
		return "InvalidAddress"
	case ConnectErrorConnectionFailure:
		return "ConnectionFailure"
	default:
		return "ConnectError"
	}
}

type SendError uint

const (
	SendErrorNotConnected = SendError(iota + 1)
	SendErrorWriteFailure
	SendErrorContextCanceled
	SendErrorDeadlineExceeded
)

func (self SendError) String() string {
	switch self {
	case SendErrorNotConnected:
		return "NotConnected"
	case SendErrorWriteFailure:
		return "WriteFailure"
	case SendErrorContextCanceled:
		return "ContextCanceled"
	case SendErrorDeadlineExceeded:
		return "DeadlineExceeded"
	default:
		return "SendError"
	}
}

type ReceiveError uint

const (
	ReceiveErrorNotConnected = ReceiveError(iota + 1)
	ReceiveErrorReadFailure
	ReceiveErrorContextCanceled
	ReceiveErrorDeadlineExceeded
	ReceiveErrorEOF
)

func (self ReceiveError) String() string {
	switch self {
	case ReceiveErrorNotConnected:
		return "NotConnected"
	case ReceiveErrorReadFailure:
		return "ReadFailure"
	case ReceiveErrorContextCanceled:
		return "ContextCanceled"
	case ReceiveErrorDeadlineExceeded:
		return "DeadlineExceeded"
	case ReceiveErrorEOF:
		return "EOF"
	default:
		return "ReceiveError"
	}
}
