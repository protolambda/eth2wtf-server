package msgtyp

type MsgTypeID byte

const (
	EventIndexUpdate MsgTypeID = iota + 1
	EventRangeRequest
	// TODO: more types
)
