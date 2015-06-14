package server

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/samalba/skyproxy/common"
)

// Receiver handles the lifetime of a Receiver connecting to the Server
type Receiver struct {
	conn   *common.ClientConn
	header common.ClientHeader
}

func (r *Receiver) readHeader() error {
	var rawHeader []byte
	// The header ends with "\n\n"
	for {
		buf, err := r.conn.ReadSlice('\n')
		if err != nil {
			return err
		}
		rawHeader = append(rawHeader, buf...)
		b, err := r.conn.ReadByte()
		if err != nil {
			return err
		}
		if b == '\n' {
			// We reached the end of the header
			break
		}
		// We are not at the end, add the last read byte to the buffer
		rawHeader = append(rawHeader, b)
	}
	if err := json.Unmarshal(rawHeader, &r.header); err != nil {
		return err
	}
	return nil
}

// NewReceiver creates a client context when there is a new connection
func NewReceiver(conn net.Conn) (*Receiver, error) {
	r := &Receiver{conn: common.NewClientConn(conn)}
	if err := r.readHeader(); err != nil {
		return nil, fmt.Errorf("Cannot read client header: %s", err)
	}
	return r, nil
}
