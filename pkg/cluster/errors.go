package cluster

type SetError uint

const (
	SetErrorKeyNotInOwnedHashSlot = SetError(iota + 1)
	SetErrorBlankKey
)

func (self SetError) String() string {
	switch self {
	case SetErrorKeyNotInOwnedHashSlot:
		return "KeyNotInOwnedHashSlot"
	case SetErrorBlankKey:
		return "BlankKey"
	default:
		return "SetError"
	}
}

type GetError uint

const (
	GetErrorKeyNotInOwnedHashSlot = GetError(iota + 1)
	GetErrorBlankKey
	GetErrorKeyNotFound
)

func (self GetError) String() string {
	switch self {
	case GetErrorKeyNotInOwnedHashSlot:
		return "KeyNotInOwnedHashSlot"
	case GetErrorBlankKey:
		return "BlankKey"
	case GetErrorKeyNotFound:
		return "KeyNotFound"
	default:
		return "GetError"
	}
}

type DeleteError uint

const (
	DeleteErrorKeyNotInOwnedHashSlot = DeleteError(iota + 1)
	DeleteErrorBlankKey
)

func (self DeleteError) String() string {
	switch self {
	case DeleteErrorKeyNotInOwnedHashSlot:
		return "KeyNotInOwnedHashSlot"
	case DeleteErrorBlankKey:
		return "BlankKey"
	default:
		return "DeleteError"
	}
}

type KeyOwnerError uint

const (
	KeyOwnerErrorBlankKey = KeyOwnerError(iota + 1)
	KeyOwnerErrorMissingOwner
)

func (self KeyOwnerError) String() string {
	switch self {
	case KeyOwnerErrorBlankKey:
		return "BlankKey"
	case KeyOwnerErrorMissingOwner:
		return "MissingOwner"
	default:
		return "KeyOwnerError"
	}
}
