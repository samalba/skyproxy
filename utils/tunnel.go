package utils

import (
	"io"
	"log"
	"net"
	"sync"
)

// TunnelConn is a low level function which takes two connections and tunnel
// one to the other. It also handles the traffic back.
func TunnelConn(from, to net.Conn, closeConns bool) {
	if closeConns {
		defer from.Close()
		defer to.Close()
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		// Receive traffic back
		nWrittenBytes, err := io.Copy(to, from)
		if err != nil {
			log.Printf("TunnelConn: cannot get traffic back: %s", err)
			return
		}
		log.Printf("TunnelConn: %d bytes received", nWrittenBytes)
	}()
	go func() {
		defer wg.Done()
		// Sending traffic to the tunnel
		nWrittenBytes, err := io.Copy(from, to)
		if err != nil {
			log.Printf("TunnelConn: cannot send traffic %s", err)
			return
		}
		log.Printf("TunnelConn: %d bytes sent", nWrittenBytes)
	}()
	wg.Wait()
}
