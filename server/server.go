package server

import (
	"log"
	"net"
)

type Server struct {
	ListenAddress string
}

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
				log.Println(err)
				return
			}
			client.Serve()
		}()
	}
}
