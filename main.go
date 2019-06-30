package main

import (
	"fmt"
	"github.com/bp1222/mud/game"
	"os"
	"strconv"
)

func main() {
	var copyoverFiles = 0
	var err error
	server := game.NewServer()

	fmt.Println("Welcome to Game")

	if len(os.Args) > 1 {
		if copyoverFiles, err = strconv.Atoi(os.Args[1]); err != nil {
			copyoverFiles = 0
		}
	}

	server.Run(copyoverFiles)

	fmt.Println("Exiting")
}
