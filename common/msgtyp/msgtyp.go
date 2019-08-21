package msgtyp

type MsgTypeID byte

const (
	ChunkRequest MsgTypeID = iota + 1
	ViewportUpdate
	// TODO: more types
)
