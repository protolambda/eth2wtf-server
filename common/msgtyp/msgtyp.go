package msgtyp

type MsgTypeID byte

const (
	EventRangeRequest MsgTypeID = iota + 1
	EventIndexUpdate
	// TODO: more types
)
