package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/samalba/skyproxy/utils"
)

// Client handles the client connection
type Client struct {
	HTTPHost string
	tcpConn  net.Conn
	tlsConn  *tls.Conn
}

// TLSConfig is used by the HTTP client
type TLSConfig struct {
	CAFile string
}

// Connect to a skyproxy server
func (c *Client) Connect(address string, tlsConfig *TLSConfig) error {
	var (
		err        error
		httpClient *httputil.ClientConn
	)
	if tlsConfig != nil {
		var conn *tls.Conn
		conn, err = c.connectTLS(address, tlsConfig)
		httpClient = httputil.NewClientConn(conn, nil)
		c.tlsConn = conn
	} else {
		var conn net.Conn
		conn, err = c.connect(address)
		httpClient = httputil.NewClientConn(conn, nil)
		c.tcpConn = conn
	}
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", "/_skyproxy/register", nil)
	if err != nil {
		return err
	}
	req.Host = c.HTTPHost
	req.Header.Add("X-Skyproxy-Client-Version", "0.1")
	err = httpClient.Write(req)
	if err != nil {
		httpClient.Close()
		return err
	}
	return nil
}

func (c *Client) connectTLS(address string, tlsConfig *TLSConfig) (*tls.Conn, error) {
	roots := x509.NewCertPool()
	certData, err := ioutil.ReadFile(tlsConfig.CAFile)
	if err != nil {
		return nil, err
	}
	if ok := roots.AppendCertsFromPEM(certData); ok != true {
		return nil, fmt.Errorf("Cannot read parse CA certificate")
	}
	return tls.Dial("tcp", address, &tls.Config{RootCAs: roots})
}

func (c *Client) connect(address string) (net.Conn, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return conn, err
	}
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}
	return conn, nil
}

// Tunnel listens to the Server conn and forward all request to the Receiver
func (c *Client) Tunnel(address string) {
	var (
		err     error
		session *yamux.Session
	)
	if c.tlsConn != nil {
		session, err = yamux.Server(c.tlsConn, nil)
	} else {
		session, err = yamux.Server(c.tcpConn, nil)
	}
	if err != nil {
		log.Printf("Cannot init Yamux Server session: %s", err)
		return
	}
	for {
		stream, err := session.Accept()
		if err != nil {
			log.Printf("Cannot accept a new Yamux stream: %s. The server might have stopped responding.", err)
			return
		}
		conn, err := net.Dial("tcp", address)
		if err != nil {
			log.Printf("Cannot connect to receiver: %s", err)
			stream.Close()
			continue
		}
		utils.TunnelConn(stream, conn, true)
	}
}
