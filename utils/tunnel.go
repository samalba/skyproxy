package utils

import (
	"io"
	"log"
	"net"
)

func TunnelConn(from, to net.Conn, closeConns bool) {
	if closeConns {
		defer from.Close()
		defer to.Close()
	}
	var wg sync.WaitGroup
	wg.Add(1)
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
	// Sending traffic to the tunnel
	nWrittenBytes, err := io.Copy(from, to)
	if err != nil {
		log.Printf("TunnelConn: cannot send traffic %s", err)
		return
	}
	log.Printf("TunnelConn: %d bytes sent", nWrittenBytes)
	wg.Wait()
}
