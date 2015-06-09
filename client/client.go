package client

import (
	"net"

	"github.com/samalba/skyproxy/common"
)

// Client handles the client connection
type Client struct {
	HTTPHost string
	conn     *common.ClientConn
}

func (c *Client) sendHeader() {
	header := common.ClientHeader{}
	header.FormatVersion = 1
	header.HTTPHost = c.HTTPHost
}

// Connect to a skyproxy server
func (c *Client) Connect(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	c.conn = common.NewClientConn(conn)
	c.sendHeader()
	return nil
}

// ConnectHTTPReceiver connects to a local receiver
func (c *Client) ConnectHTTPReceiver() {
}
