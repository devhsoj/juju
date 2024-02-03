package juju

type Command = byte

const (
	PublishCommand   = 'p'
	SubscribeCommand = 's'
)

const (
	CommandTypeBufferSize          int = 1
	CommandDataLengthBufferSize    int = 8
	CommandIdentifierBufferSize    int = 128
	CommandMetadataTotalBufferSize     = CommandTypeBufferSize + CommandDataLengthBufferSize + CommandIdentifierBufferSize
)

func EncodeIdentifier(identifier string) []byte {
	if len(identifier) >= CommandIdentifierBufferSize {
		return []byte(identifier[:CommandIdentifierBufferSize])
	}

	buf := make([]byte, CommandIdentifierBufferSize)

	for i := 0; i < 128; i++ {
		if i >= len(identifier) {
			break
		}

		buf[i] = identifier[i]
	}

	return buf
}

func DecodeIdentifierBuffer(identifierBuffer []byte) string {
	if len(identifierBuffer) > CommandIdentifierBufferSize {
		identifierBuffer = identifierBuffer[:CommandIdentifierBufferSize]
	}

	var endIndex = 0

	for i := 0; i < len(identifierBuffer); i++ {
		if identifierBuffer[i] == 0 {
			endIndex = i
			break
		}
	}

	return string(identifierBuffer[:endIndex])
}
