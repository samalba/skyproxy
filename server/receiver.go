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
	buf, err := searchString(r.conn, "\n\n")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(buf, &r.header); err != nil {
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
