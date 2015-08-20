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
	id := time.Now().Nanosecond()
	var wg sync.WaitGroup
	tunnelCopy := func(label string, from, to net.Conn) {
		defer wg.Done()
		if closeConns {
			defer from.Close()
			defer to.Close()
		}
		log.Printf("TunnelConn(%d),%s: open", id, label)
		nWrittenBytes, err := io.Copy(from, to)
		if err != nil {
			log.Printf("TunnelConn(%d),%s: %s", id, label, err)
		} else {
			log.Printf("TunnelConn(%d),%s: %d bytes copied", id, label, nWrittenBytes)
		}
	}
	wg.Add(2)
	go tunnelCopy("receive", to, from)
	go tunnelCopy("send", from, to)
	wg.Wait()
}
