package client

import (
	"encoding/json"
	"io"
	"net"
	"os"

	"github.com/samalba/skyproxy/common"
)

// Client handles the client connection
type Client struct {
	HTTPHost   string
	serverConn *common.ClientConn
}

func (c *Client) sendHeader() error {
	header := &common.ClientHeader{}
	header.FormatVersion = 1
	header.Protocol = "http"
	header.HTTPHost = c.HTTPHost
	buf, err := json.Marshal(header)
	if err != nil {
		return err
	}
	buf = append(buf, []byte("\n\n")...)
	_, err = c.serverConn.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

// Connect to a skyproxy server
func (c *Client) Connect(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	c.serverConn = common.NewClientConn(conn)
	if err := c.sendHeader(); err != nil {
		return err
	}
	return nil
}

// ConnectHTTPReceiver connects to a local receiver
func (c *Client) ConnectHTTPReceiver(address string) {
}

// Forward reads data from the Server and transfer them to the Receiver
func (c *Client) Forward() {
	io.Copy(c.serverConn, os.Stdout)
}
