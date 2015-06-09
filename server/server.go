package server

import (
	"log"
	"net"
)

// Server context
type Server struct {
	ListenAddress string
}

// ListenForClients is the main loop for accepting new clients
func (s *Server) ListenForClients() error {
	listen, err := net.Listen("tcp", s.ListenAddress)
	if err != nil {
		return err
	}
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go func() {
			client, err := NewClient(conn)
			if err != nil {
				// Client gets dropped
				log.Println(err)
				return
			}
			client.Serve()
		}()
	}
}
