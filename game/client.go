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
		ParseInput()
		Close()
	}

	ClientImpl struct {
		server   Server
		conn     *net.TCPConn
		commands []Command
		reader   *bufio.Reader
	}
)

func NewClient(conn *net.TCPConn, server Server) Client {
	return &ClientImpl{
		server: server,
		conn: conn,
		reader: bufio.NewReader(conn),
	}
}

func (c *ClientImpl) Write(b []byte) {
	_, err := c.conn.Write(b)
	if err != nil {
		log.Printf("error writing to client %v", err)
	}
}

func (c *ClientImpl) WriteString(s string) {
	c.Write([]byte(s))
}

func (c *ClientImpl) ParseInput() {
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
			Commands[args[0]].Act(c)
		} else {
			c.WriteString("unknown command")
		}
	}
}

func (c *ClientImpl) Close() {
	err := c.conn.Close()
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			log.Printf("error closing client %v", err)
		}
		return
	}

	log.Printf("disconnected connection from %v", c.conn.RemoteAddr())
}
