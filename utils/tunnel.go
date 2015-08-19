package utils

import (
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// TunnelConn is a low level function which takes two connections and tunnel
// one to the other. It also handles the traffic back.
func TunnelConn(from, to net.Conn, closeConns bool) {
	if closeConns {
		defer from.Close()
		defer to.Close()
	}
	id := time.Now().Nanosecond()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		// Receive traffic back
		log.Printf("TunnelConn(%d): from <- to", id)
		nWrittenBytes, err := io.Copy(from, to)
		if err != nil {
			log.Printf("TunnelConn(%d): cannot get traffic back: %s", id, err)
			return
		}
		log.Printf("TunnelConn(%d): %d bytes received", id, nWrittenBytes)
	}()
	go func() {
		defer wg.Done()
		// Sending traffic to the tunnel
		log.Printf("TunnelConn(%d): from -> to", id)
		nWrittenBytes, err := io.Copy(to, from)
		if err != nil {
			log.Printf("TunnelConn(%d): cannot send traffic %s", id, err)
			return
		}
		log.Printf("TunnelConn(%d): %d bytes sent", id, nWrittenBytes)
	}()
	wg.Wait()
}
