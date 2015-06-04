package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
)

const (
	// If a header goes beyond this size, the client gets dropped
	maxHeaderSize = 4096
	// Marks the end of a header
	headerSeparator = "\n\n"
)

// Client handles the lifetime of a Client connecting to the Server
type Client struct {
	conn   *ClientConn
	header map[string]string
}

func (c *Client) readHeader() error {
	rawHeader, err := c.conn.Peek(maxHeaderSize)
	if err != nil {
		return err
	}
	index := bytes.Index(rawHeader, []byte("\n\n"))
	if index < 0 {
		return fmt.Errorf("Cannot reach the final size of a header")
	}
	index += 2
	size, err := c.conn.Read(rawHeader[:index])
	if err != nil {
		return err
	}
	if size < index {
		return fmt.Errorf("Cannot read full header")
	}
	c.header = make(map[string]string)
	if err := json.Unmarshal(rawHeader, &c.header); err != nil {
		return err
	}
	return nil
}

func NewClient(conn net.Conn) (*Client, error) {
	client := &Client{conn: NewClientConn(conn)}
	client.readHeader()
	return client, nil
}

func (c *Client) Serve() {
}
