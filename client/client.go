package client

import (
	"encoding/json"
	"io"
	"log"
	"net"

	"github.com/samalba/skyproxy/common"
)

// Client handles the client connection
type Client struct {
	HTTPHost     string
	serverConn   *common.ClientConn
	receiverConn *common.ClientConn
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
func (c *Client) ConnectHTTPReceiver(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	c.receiverConn = common.NewClientConn(conn)
	return nil
}

// Forward reads data from the Skyproxy server and send it to the receiver (and vice versa)
func (c *Client) Forward() {
	go func() {
		// Sending traffic back from the Receiver to the Skyproxy server
		nWrittenBytes, err := io.Copy(c.receiverConn, c.serverConn)
		if err != nil {
			log.Printf("[client] Cannot forward data from the Receiver to the Skyproxy server: %s", err)
			return
		}
		log.Printf("[client] Receiver -> Skyproxy server: %d bytes", nWrittenBytes)
	}()
	// Sending traffic from the Skyproxy server to the receiver
	nWrittenBytes, err := io.Copy(c.serverConn, c.receiverConn)
	if err != nil {
		log.Printf("[client] Cannot forward data from the Skyproxy server to the Receiver: %s", err)
		return
	}
	log.Printf("[client] Skyproxy server -> Receiver: %d bytes", nWrittenBytes)
}
