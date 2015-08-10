package client

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/samalba/skyproxy/utils"
)

// Client handles the client connection
type Client struct {
	HTTPHost   string
	serverConn net.Conn
}

// Connect to a skyproxy server
func (c *Client) Connect(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	httpClient := httputil.NewClientConn(conn, nil)
	req, err := http.NewRequest("POST", "/_skyproxy/register", nil)
	if err != nil {
		return err
	}
	req.Header.Add("Host", c.HTTPHost)
	req.Header.Add("X-Skyproxy-Client-Version", "0.1")
	err = httpClient.Write(req)
	if err != nil {
		httpClient.Close()
		return err
	}
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}
	c.serverConn = conn
	return nil
}

// Tunnel listens to the Server conn and forward all request to the Receiver
func (c *Client) Tunnel(address string) {
	//FIXME(samalba): handle multi-plexing: concurrent requests coming in. Right
	// now they would be mixed together. Each request should be de-multi-plexed
	// in a single socket opened against the receiver.
	for {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			log.Printf("Cannot connect to receiver: %s", err)
			time.Sleep(3 * time.Second)
			continue
		}
		utils.TunnelConn(c.serverConn, conn, false)
		conn.Close()
	}
}
