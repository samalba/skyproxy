package server

import (
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/samalba/skyproxy/utils"
)

// Client context
// each struct represents a skyproxy Client (with the HTTPHost it registered)
type Client struct {
	Conn     net.Conn
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
						s.clientList[host] = append(l[:i], l[i+1:]...)
						log.Printf("Client unregistered for HTTP host: %s", host)
						break
					}
				}
				if len(l) == 0 {
					delete(s.clientList, host)
					log.Printf("Removed HTTP host: %s", host)
				}
			}
			s.numClients--
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
		s.clientIn <- &Client{Conn: conn, HTTPHost: r.Host}
	}
	return h
}

// createPublicHTTPHandler returns the handler that manages the Public HTTP traffic
func createPublicHTTPHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	h := func(w http.ResponseWriter, r *http.Request) {
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
		clientList, exists := s.clientList[r.Host]
		if !exists {
			http.Error(w, "There is no Receiver registered for this Host", http.StatusInternalServerError)
			log.Printf("Cannot handle request for Host %s: no receiver registed for this Host", r.Host)
			return
		}
		// Pick a client randomly
		idx := s.random.Intn(len(clientList))
		client := clientList[idx]
		utils.TunnelConn(conn, client.Conn, false)
		// FIXME: close connections and handle errors
	}
	return h
}

// StartHTTPServer creates an HTTP server
func (s *Server) StartHTTPServer(address string) error {
	// Start the routine to manage the clients (in & out)
	go s.manageClientList()
	// Start the HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/_skyproxy/register", createClientsHTTPHandler(s))
	mux.HandleFunc("/", createPublicHTTPHandler(s))
	return http.ListenAndServe(address, mux)
}
