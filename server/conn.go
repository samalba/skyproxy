package server

import (
	"bufio"
	"net"
)

// ClientConn wraps is a buffered net.Conn
type ClientConn struct {
	net.Conn // embeds net.Conn
	bufr     *bufio.Reader
}

// NewClientConn inits a ClientConn from a net.Conn
func NewClientConn(conn net.Conn) *ClientConn {
	// NOTE(samalba): in case we need to tweak the size of the buffer,
	// use: bufio.NewReaderSize() instead.
	return &ClientConn{conn, bufio.NewReader(conn)}
}

// Peek grabs data on the buffer without consuming it
func (c *ClientConn) Peek(n int) ([]byte, error) {
	return c.bufr.Peek(n)
}

// Read reads from the buffer
func (c *ClientConn) Read(p []byte) (int, error) {
	return c.bufr.Read(p)
}
