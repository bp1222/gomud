package game

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type (
	Command interface {
		Act(Client, []string)
	}

	copyover struct {}
	echo struct {}
	quit struct {}
)

var Commands = map[string]Command {
	"copyover": &copyover{},
	"echo": &echo{},
	"quit": &quit{},
}

func (c *copyover) Act(client Client, args []string) {
	var extraFiles []*os.File
	server := client.Server()

	// Push the server in first
	serverFile, err := server.Listener().File()
	if err != nil {
		log.Printf("copyover aborted, unable to create server file %v", err)
		return
	}
	extraFiles = append(extraFiles, serverFile)

	// Create copies of the clients to pass into new program
	server.Clients().Foreach(func(c interface{}) {
		client := c.(Client)

		clientFile, err := client.Conn().File()
		if err != nil {
			return
		}

		client.Write([]byte("*** copyover started, hold on ***"))
		extraFiles = append(extraFiles, clientFile)
	})

	// Need to get fd's for server, and clients to pass on
	cmd := exec.Command("./mud", fmt.Sprint(len(extraFiles)))
	cmd.ExtraFiles = extraFiles
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Alert server it's time to die
	server.Stop()

	// Run the new server, see you on the other side
	if err := cmd.Start(); err != nil {
		log.Printf("copyover failed %v", err)
	}
}

func (c *echo) Act(client Client, args []string) {
	if len(args) >= 1 {
		client.WriteString(strings.Join(args, " ") + "\n")
	}
}

func (c *quit) Act(client Client, args []string) {
	client.WriteString("You have left\n")
	client.Close()
}

