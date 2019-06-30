package game

import (
	"bufio"
	"log"
	"net"
	"strings"
)

type (
	Client interface {
		Write([]byte)
		WriteString(string)
		ParseInput() error
		Server() Server
		Conn() *net.TCPConn
		Close()
	}

	clientImpl struct {
		server   Server
		conn     *net.TCPConn
		reader   *bufio.Reader
	}
)

func NewClient(conn *net.TCPConn, server Server) Client {
	return &clientImpl{
		server: server,
		conn: conn,
		reader: bufio.NewReader(conn),
	}
}

func (c *clientImpl) Conn() *net.TCPConn {
	return c.conn
}

func (c *clientImpl) Server() Server {
	return c.server
}

func (c *clientImpl) Write(b []byte) {
	_, err := c.conn.Write(b)
	if err != nil {
		log.Printf("error writing to client %v", err)
	}
}

func (c *clientImpl) WriteString(s string) {
	c.Write([]byte(s))
}

func (c *clientImpl) ParseInput() (err error) {
	input, err := c.reader.ReadString('\n')
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			log.Printf("error reding client string %v", err)
		}
		return
	}

	log.Printf("input received %v", input)

	args := strings.Split(input, " ")
	for i := range args {
		args[i] = strings.TrimSpace(args[i])
	}
	if len(args) > 0 {
		if Commands[args[0]] != nil {
			Commands[args[0]].Act(c, args[1:])
		} else {
			c.WriteString("unknown command")
		}
	}

	return
}

func (c *clientImpl) Close() {
	err := c.conn.Close()
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			log.Printf("error closing client %v", err)
		}
		return
	}

	log.Printf("disconnected connection from %v", c.conn.RemoteAddr())
}
