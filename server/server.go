package server

import (
	"http"
	"io"
	"log"
	"math/rand"
	"net"
	"sync"
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
	ClientsHTTPAddress string
	PublicHTTPAddress  string
	clientList         map[string][]*Client
	numClients         int
	clientIn           chan *Client
	clientOut          chan *Client
	random             *rand.Rand
}

// NewServer is usually called once to create the server context
func NewServer(address, httpAddress string) *Server {
	s := &Server{ListenAddress: address, ListenHTTPAddress: httpAddress}
	s.clientList = make(map[string][]*Client)
	s.numReceivers = 0
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

// ListenForClients is the main loop for accepting new clients
func (s *Server) ListenForClients() error {
	// httpHandler handles request coming from public HTTP traffic
	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
			return
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		host := r.Header.Get("Host")
		if host == "" {
			http.Error(w, "No Host header specified", http.StatusBadRequest)
			return
		}
		s.clientIn <- &Client{Conn: conn, HTTPHost: host}
	}
	// Inits the http server
	mux := http.NewServeMux()
	mux.HandleFunc("/", httpHandler)
	http.ListenAndServe(s.ClientsHTTPAddress, mux)
}

// ListenForHTTP listens for HTTP traffic
func (s *Server) ListenForHTTP() error {
	// httpHandler handles request coming from public HTTP traffic
	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
			return
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		clientList, exists := s.clientList[httpClient.HTTPHost]
		if !exists {
			http.Error(w, "There is no Receiver registered for this Host", http.StatusInternalServerError)
			return
		}
		// Pick a client randomly
		idx := s.random.Intn(len(clientList))
		client := clientList[idx]
		utils.TunnelConn(conn, client.Conn, false)
		// FIXME: close connections and handle errors
	}
	// Inits the http Server
	mux := http.NewServeMux()
	mux.HandleFunc("/", httpHandler)
	http.ListenAndServe(s.PublicHTTPAddress, mux)
}
