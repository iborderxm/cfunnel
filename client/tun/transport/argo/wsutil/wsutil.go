package wsutil

import (
	"io"
)

func ReadClientBinary(rw io.Reader) ([]byte, error) {
	buf := make([]byte, 4096)
	n, err := rw.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func WriteServerBinary(wr io.Writer, data []byte) error {
	_, err := wr.Write(data)
	return err
}

func WriteServerMessage(wr io.Writer, messageType int, data []byte) error {
	_, err := wr.Write(data)
	return err
}

func HandleClientControlMessage(rw io.ReadWriter, msg Message) error {
	return nil
}

type Message struct {
	OpCode  int
	Payload []byte
}
