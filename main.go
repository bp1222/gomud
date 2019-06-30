package main

import (
	"fmt"
	"github.com/bp1222/mud/game"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	sg := make(chan os.Signal)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)

	game.GameServer = game.NewServer()

	fmt.Println("Welcome to Game")

	game.GameServer.Start()
	for {
		if !game.GameServer.Running() {
			break
		}

		select {
		case <-sg:
			game.GameServer.Stop()
		case <-time.After(time.Millisecond * 100):
			continue
		}
	}
	game.GameServer.Stop()

	fmt.Println("Exiting")
}
