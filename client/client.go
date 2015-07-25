package client

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/samalba/skyproxy/common"
)

// Client handles the client connection
type Client struct {
	HTTPHost     string
	ReceiverAddr string
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

// connectHTTPReceiver connects to a local receiver
func (c *Client) connectHTTPReceiver(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	c.receiverConn = common.NewClientConn(conn)
	return nil
}

func (c *Client) reconnectHTTPReceiver() {
	if c.receiverConn != nil {
		c.receiverConn.Close()
	}
	for {
		conn, err := net.Dial("tcp", c.ReceiverAddr)
		if err != nil {
			log.Printf("[client] Cannot connect to %s: %s. Retrying in 1 second", c.ReceiverAddr, err)
			time.Sleep(1 * time.Second)
			continue
		}
		c.receiverConn = common.NewClientConn(conn)
		break
	}
	log.Printf("[client] Connected to HTTP receiver %s", c.ReceiverAddr)
}

// Forward reads data from the Skyproxy server and send it to the receiver (and vice versa)
func (c *Client) Forward() {
	var wg sync.WaitGroup
	for {
		c.reconnectHTTPReceiver()
		wg.Add(1)
		go func() {
			// Sending traffic back from the Receiver to the Skyproxy server
			defer wg.Done()
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
		} else {
			log.Printf("[client] Skyproxy server -> Receiver: %d bytes", nWrittenBytes)
		}
		wg.Wait()
	}
}
