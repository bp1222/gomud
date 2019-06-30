package game

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type (
	Server interface {
		Run(int)
		Stop()
		Running() bool

		Clients() SafeSlice
		Listener() *net.TCPListener

		handleClient(client Client)
		disconnectClient(client Client)
	}

	ServerImpl struct {
		listener *net.TCPListener
		clients  SafeSlice
		ticker   *time.Ticker
		shutdown bool
		stop chan bool
	}
)

const (
	gameTime = time.Millisecond * 100
)

func NewServer() Server {
	return &ServerImpl {
		ticker:   time.NewTicker(gameTime),
		clients:  NewSafeSlice(),
		shutdown: false,
		stop: make(chan bool),
	}
}

func (s *ServerImpl) Listener() *net.TCPListener {
	return s.listener
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
		if err := client.ParseInput(); err != nil {
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

func (s *ServerImpl) Run(copyoverFiles int) {
	var err error

	if copyoverFiles == 0 {
		s.listener, err = net.ListenTCP("tcp4", &net.TCPAddr{Port: 8088})
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				log.Printf("error listening for connection %v", err)
			}
			return
		}
	} else {
		// Copyover, need to recover
		var ok bool
		var fd uintptr = 3

		// Allow some time
		time.Sleep(time.Second * 2)

		serverFile := os.NewFile(fd, "server")
		newListener, err := net.FileListener(serverFile)

		if err := serverFile.Close(); err != nil {
			log.Printf("unable to close incooming file")
		}

		if err != nil {
			log.Printf("unable to reconnect to server %v", err)
			return
		}

		s.listener, ok = newListener.(*net.TCPListener)
		if !ok {
			log.Println("failed to start after copyover")
			return
		}

		// Load Clients Back
		for i := 1; i < copyoverFiles; i++ {
			fd += 1
			clientFile := os.NewFile(fd, "client")
			newConn, err := net.FileConn(clientFile)

			if err := clientFile.Close(); err != nil {
				log.Printf("unable to close incooming file")
			}

			if err != nil {
				if err := clientFile.Close(); err != nil {
					log.Printf("client was lost in copyover")
				}
				continue
			}

			newConn, ok = newConn.(*net.TCPConn)
			if !ok {
				log.Printf("client was lost in copyover")
				continue
			}

			if _, err := newConn.Write([]byte("\n\nWelcome Back From Copyover\n")); err != nil {
				log.Printf("couldn't welcome back from copyover")
				_ = newConn.Close()
				continue
			}

			go s.handleClient(NewClient(newConn.(*net.TCPConn), s))
		}
	}

	// Start routine to listen for new connections
	go s.start()

	// Run the game loop
	sg := make(chan os.Signal)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)

	select {
	case <-sg:
		s.stopServer()
		break
	case <-s.stop:
		fmt.Println("Got Request To Die")
		s.stopServer()
		break
	}
}

func (s *ServerImpl) Running() bool {
	return !s.shutdown
}

func (s *ServerImpl) stopServer() {
	// Denote server is shutting down
	s.shutdown = true

	fmt.Println("Killing Clients")
	// Kill clients
	s.clients.Foreach(func(client interface{}) {
		client.(Client).Close()
	})

	fmt.Println("Waiting On Clients To Die")
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

func (s *ServerImpl) Stop() {
	// Alert the game loop to die
	s.stop <- true
}
