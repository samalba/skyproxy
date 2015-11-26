package server

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/samalba/skyproxy/utils"
)

// Client context
// each struct represents a skyproxy Client (with the HTTPHost it registered)
type Client struct {
	Conn     net.Conn
	Session  *yamux.Session
	HTTPHost string
}

// Server context
type Server struct {
	clientList map[string][]*Client
	numClients int
	clientIn   chan *Client
	clientOut  chan *Client
	random     *rand.Rand
}

// TLSConfig is used by the HTTP server
type TLSConfig struct {
	CertFile string
	KeyFile  string
}

// NewServer is usually called once to create the server context
func NewServer() *Server {
	s := &Server{}
	s.clientList = make(map[string][]*Client)
	s.numClients = 0
	s.clientIn = make(chan *Client, 10)
	s.clientOut = make(chan *Client, 10)
	// init rand seed
	s.random = rand.New(rand.NewSource(time.Now().UnixNano()))
	return s
}

// manageClientList maps Client add/del to the connect/disconnect of clients
func (s *Server) manageClientList() {
	for {
		var client *Client
		select {
		// New client
		case client = <-s.clientIn:
			host := client.HTTPHost
			if l, exists := s.clientList[host]; exists {
				// There is already one or several clients for this Hostname
				s.clientList[host] = append(l, client)
			} else {
				s.clientList[host] = []*Client{client}
				log.Printf("New HTTP host: %s", host)
			}
			s.numClients++
			log.Printf("New client registered for HTTP host: %s", host)
		// Removing client
		case client = <-s.clientOut:
			host := client.HTTPHost
			if l, exists := s.clientList[host]; exists {
				for i, c := range l {
					if c == client {
						// Removing client from the list
						defer c.Session.Close()
						defer c.Conn.Close()
						s.clientList[host] = append(l[:i], l[i+1:]...)
						log.Printf("Client unregistered for HTTP host: %s", host)
						s.numClients--
						break
					}
				}
				if len(l) == 0 {
					delete(s.clientList, host)
					log.Printf("Removed HTTP host: %s", host)
				}
			}
		}
		log.Printf("%d clients connected", s.numClients)
	}
}

// createClientsHTTPHandler returns the handler that manages Skyproxy Clients
func createClientsHTTPHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	h := func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
			log.Println("Cannot register new client: hijacking not supported")
			return
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("Cannot register new client: %s", err)
			return
		}
		if r.Host == "" {
			http.Error(w, "No Host header specified", http.StatusBadRequest)
			log.Println("Cannot register new client: no host header specified")
			return
		}
		session, err := yamux.Client(conn, nil)
		if err != nil {
			http.Error(w, "Cannot init Yamux Client session", http.StatusInternalServerError)
			log.Printf("Cannot init Yamux Client session: %s", err)
			return
		}
		s.clientIn <- &Client{Conn: conn, Session: session, HTTPHost: r.Host}
	}
	return h
}

func (s *Server) pickRandomClientStream(host string) (*yamux.Stream, error) {
	for retry := 0; retry < 5; retry++ {
		clientList, exists := s.clientList[host]
		if !exists {
			return nil, fmt.Errorf("Cannot handle request for Host %s: no Client registered for this Host", host)
		}
		// Pick a client randomly
		idx := s.random.Intn(len(clientList))
		client := clientList[idx]
		stream, err := client.Session.OpenStream()
		if err != nil {
			log.Printf("Cannot open a new Yamux stream on the Client session: %s", err)
			s.clientOut <- client
			continue
		}
		return stream, nil
	}
	return nil, fmt.Errorf("Cannot find a registered Client with an active connection")
}

// createPublicHTTPHandler returns the handler that manages the Public HTTP traffic
func createPublicHTTPHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	h := func(w http.ResponseWriter, r *http.Request) {
		// Pick a random client and open a new Yamux stream
		stream, err := s.pickRandomClientStream(r.Host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("Cannot find a valid registered Client: %s", err)
			return
		}
		defer stream.Close()
		log.Printf("Found a valid client registered for Host %s", r.Host)
		// Got a valid client, hijack the connection
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
			log.Printf("Cannot handle request for Host %s: hijacking not supported", r.Host)
			return
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("Cannot handle request for Host %s: %s", r.Host, err)
			return
		}
		// Send the initial request to the client
		r.Write(stream)
		utils.TunnelConn(conn, stream, true)
	}
	return h
}

// StartServer creates an HTTP(s) server
func (s *Server) StartServer(address string, clientsManager bool, tlsConfig *TLSConfig) error {
	mux := http.NewServeMux()
	if clientsManager == true {
		// Start the routine to manage the clients (in & out)
		go s.manageClientList()
		// Register the route to manage the SkyProxy cients
		mux.HandleFunc("/_skyproxy/register", createClientsHTTPHandler(s))
	} else {
		// Register the route to handle the public HTTP(s) traffic
		mux.HandleFunc("/", createPublicHTTPHandler(s))
	}
	if tlsConfig != nil {
		return http.ListenAndServeTLS(address, tlsConfig.CertFile, tlsConfig.KeyFile, mux)
	}
	return http.ListenAndServe(address, mux)
}
