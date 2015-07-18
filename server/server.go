package server

import (
	"log"
	"math/rand"
	"net"
	"time"
)

// Server context
type Server struct {
	ListenAddress     string
	ListenHTTPAddress string
	receiverList      map[string][]*Receiver
	numReceivers      int
	receiverIn        chan *Receiver
	receiverOut       chan *Receiver
}

var random *rand.Rand

// NewServer is usually called once to create the server context
func NewServer(address, httpAddress string) *Server {
	s := &Server{ListenAddress: address, ListenHTTPAddress: httpAddress}
	s.receiverList = make(map[string][]*Receiver)
	s.numReceivers = 0
	s.receiverIn = make(chan *Receiver, 10)
	s.receiverOut = make(chan *Receiver, 10)
	// init rand seed
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
	return s
}

func (s *Server) manageReceiverList() {
	for {
		var receiver *Receiver
		select {
		// New receiver
		case receiver = <-s.receiverIn:
			host := receiver.header.HTTPHost
			if l, exists := s.receiverList[host]; exists {
				// There is already one or several receivers for this Hostname
				s.receiverList[host] = append(l, receiver)
			} else {
				s.receiverList[host] = []*Receiver{receiver}
				log.Printf("New host %s", host)
			}
			s.numReceivers++
			log.Printf("New receiver registered on %s", host)
		// Removing receiver
		case receiver = <-s.receiverOut:
			host := receiver.header.HTTPHost
			if l, exists := s.receiverList[host]; exists {
				for i, c := range l {
					if c == receiver {
						// Removing receiver from the list
						s.receiverList[host] = append(l[:i], l[i+1:]...)
						log.Printf("Receiver unregistered on %s", host)
						break
					}
				}
				if len(l) == 0 {
					delete(s.receiverList, host)
					log.Printf("Removed host %s", host)
				}
			}
			s.numReceivers--
		}
		log.Printf("%d receivers connected", s.numReceivers)
	}
}

// ListenForReceivers is the main loop for accepting new receivers
func (s *Server) ListenForReceivers() error {
	listen, err := net.Listen("tcp", s.ListenAddress)
	if err != nil {
		return err
	}
	log.Printf("Listening for receivers on %s", s.ListenAddress)
	defer listen.Close()
	go s.manageReceiverList()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		// Handles each new receiver in a routine
		go func() {
			receiver, err := NewReceiver(conn)
			if err != nil {
				// Receiver gets dropped
				log.Println(err)
				return
			}
			s.receiverIn <- receiver
		}()
	}
}

// ListenForHTTP listens for HTTP traffic
func (s *Server) ListenForHTTP() error {
	listen, err := net.Listen("tcp", s.ListenHTTPAddress)
	if err != nil {
		return err
	}
	log.Printf("Listening for HTTP on %s", s.ListenHTTPAddress)
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		httpClient, err := NewHTTPClient(conn)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("HTTP Client for host: %s", httpClient.HTTPHost)
		//FIXME: send the http header buffer to the receiver
		go s.forwardTraffic(httpClient)
	}
}

// forwardTraffic reads data from HTTP cients and send it to receivers (and vice versa)
func (s *Server) forwardTraffic(httpClient *HTTPClient) {
	receiverList, exists := s.receiverList[httpClient.HTTPHost]
	if !exists {
		log.Printf("There is no Receiver register for the HTTP host: %s", httpClient.HTTPHost)
		return
	}
	// Pick a receiver randomly
	idx := random.Intn(len(receiverList))
	receiver := receiverList[idx]
	go func() {
		var buf = make([]byte, 4096)
		// Sending traffic back from the Receiver to the HTTP Client
		for {
			log.Println("Reading from Receiver...")
			nReadBytes, err := receiver.Read(buf)
			if err != nil {
				log.Printf("Cannot read from the Receiver: %s", err)
				return
			}
			nWrittenBytes, err := httpClient.Write(buf[:nReadBytes])
			if err != nil {
				log.Printf("Cannot write to the HTTP Client: %s", err)
				return
			}
			if nReadBytes != nWrittenBytes {
				log.Printf("Unable to forward all the data from the Receiver to the HTTP Client (received %d bytes, forwarded %d bytes)", nReadBytes, nWrittenBytes)
			}
		}
	}()
	// Sending traffic from the HTTP Client to the Receiver
	var buf = make([]byte, 4096)
	for {
		log.Println("Reading from HTTP Client...")
		nReadBytes, err := httpClient.Read(buf)
		if err != nil {
			log.Printf("Cannot read from the HTTP Client: %s", err)
			return
		}
		nWrittenBytes, err := receiver.Write(buf[:nReadBytes])
		if err != nil {
			log.Printf("Cannot write to the Receiver: %s", err)
			return
		}
		if nReadBytes != nWrittenBytes {
			log.Printf("Unable to forward all the data from the HTTP Client to the Receiver (received %d bytes, forwarded %d bytes)", nReadBytes, nWrittenBytes)
		}
		log.Printf("nReadBytes = %d; nWrittenBytes = %d\n", nReadBytes, nWrittenBytes)
	}
}
