package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/samalba/skyproxy/common"
)

// Client handles the lifetime of a Client connecting to the Server
type Client struct {
	conn   *common.ClientConn
	header common.ClientHeader
}

func (c *Client) readHeader() error {
	var rawHeader []byte
	// The header ends with "\n\n"
	for {
		buf, err := c.conn.ReadSlice('\n')
		if err != nil {
			return err
		}
		rawHeader = append(rawHeader, buf...)
		b, err := c.conn.ReadByte()
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
	if err := json.Unmarshal(rawHeader, &c.header); err != nil {
		return err
	}
	return nil
}

// NewClient creates a client context when there is a new connection
func NewClient(conn net.Conn) (*Client, error) {
	client := &Client{conn: common.NewClientConn(conn)}
	if err := client.readHeader(); err != nil {
		return nil, fmt.Errorf("Cannot read client header: %s", err)
	}
	return client, nil
}

// Serve serves a new client after the intial handshake
func (c *Client) Serve() {
	log.Printf("New Client, header: %#v", c.header)
}
