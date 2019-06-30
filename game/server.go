package game

import (
	"log"
	"net"
	"time"
)

// Global Server For Game
var GameServer Server

type (
	Server interface {
		Start()
		Stop()
		Running() bool

		start()
		handleClient(client Client)
		disconnectClient(client Client)
	}

	ServerImpl struct {
		listener *net.TCPListener
		clients  SafeSlice
		ticker   *time.Ticker
		shutdown bool
	}
)

const (
	gameTime = time.Millisecond * 1000
)

func NewServer() Server {
	return &ServerImpl {
		ticker:   time.NewTicker(gameTime),
		clients:  NewSafeSlice(),
		shutdown: false,
	}
}

func (s *ServerImpl) Clients() SafeSlice {
	return s.clients
}

func (s *ServerImpl) disconnectClient(client Client) {
	client.Close()
	s.clients.Remove(client)
}

func (s *ServerImpl) handleClient(client Client) {
	s.clients.Add(client)
	defer s.disconnectClient(client)

	for {
		client.ParseInput()

		if s.shutdown {
			break
		}

		<-s.ticker.C
	}
}

func (s *ServerImpl) start() {
	for {
		conn, err := s.listener.AcceptTCP()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				log.Printf("error accepting connection %v", err)
			}
			break
		}

		log.Printf("accepted connection from %v", conn.RemoteAddr())

		go s.handleClient(NewClient(conn, s))
	}
}

func (s *ServerImpl) Start () {
	var err error

	// Start listening
	s.listener, err = net.ListenTCP("tcp4", &net.TCPAddr{Port: 8088})
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			log.Printf("error listening for connection %v", err)
		}
		return
	}

	// Start routine to listen for new connections
	go s.start()
}

func (mn *ServerImpl) Running() bool {
	return !mn.shutdown
}

func (s *ServerImpl) Stop () {
	// Denote server is shutting down
	s.shutdown = true

	// Kill clients
	s.clients.Foreach(func(client interface{}) {
		client.(Client).Close()
	})

	// Wait for clients to die
	s.clients.Wait()

	// Close us out
	err := s.listener.Close()
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			log.Printf("error closing listener %v", err)
		}
	}
}
