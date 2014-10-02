package service

type Message interface {
	Type() MessageType
	Source() string
}

type MessageType uint8

const (
	MessageTypeError MessageType = iota
)

type MessageError struct {
	source string
	Error  error
}

func (me MessageError) Type() MessageType {
	return MessageTypeError
}

func (me MessageError) Source() string {
	return me.source
}
