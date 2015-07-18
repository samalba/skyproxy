package server

import (
	"fmt"
	"net"
	"strings"

	"github.com/samalba/skyproxy/common"
)

// HTTPClient is the context of a client making an HTTP request
type HTTPClient struct {
	conn         *common.ClientConn
	headerBuffer []byte
	HTTPHost     string
}

// searchString searches for a pattern inside a buffer and returns the consumed buffer
func searchString(bufr *common.ClientConn, pattern string) ([]byte, error) {
	var read []byte
	for {
		found := true
		buf, err := bufr.ReadSlice(pattern[0])
		if err != nil {
			return nil, err
		}
		read = append(read, buf...)
		for _, c := range pattern[1:] {
			b, err := bufr.ReadByte()
			if err != nil {
				return nil, err
			}
			read = append(read, b)
			if b != byte(c) {
				found = false
				break
			}
		}
		if found {
			// We found the pattern
			return read, nil
		}
	}
}

func (c *HTTPClient) readHTTPHost() error {
	buf, err := searchString(c.conn, "Host: ")
	if err != nil {
		return err
	}
	c.headerBuffer = append(c.headerBuffer, buf...)
	buf, err = searchString(c.conn, "\n")
	if err != nil {
		return err
	}
	c.headerBuffer = append(c.headerBuffer, buf...)
	c.HTTPHost = strings.TrimRight(string(buf), "\r\n")
	return nil
}

// NewHTTPClient creates an HTTPClient from a net.Conn
func NewHTTPClient(conn net.Conn) (*HTTPClient, error) {
	c := &HTTPClient{conn: common.NewClientConn(conn)}
	if err := c.readHTTPHost(); err != nil {
		return nil, fmt.Errorf("Cannot parse client host header: %s", err)
	}
	return c, nil
}

// Read reads data from the HTTP client socket
func (c *HTTPClient) Read(buf []byte) (int, error) {
	if len(c.headerBuffer) > 0 {
		// There are remaining bytes from the header parsing, return them instead of reading more data
		n := copy(buf, c.headerBuffer)
		c.headerBuffer = c.headerBuffer[n:]
		return n, nil
	}
	return c.conn.Read(buf)
}

// Write writes data to the HTTP client socket
func (c *HTTPClient) Write(buf []byte) (int, error) {
	return c.conn.Write(buf)
}
