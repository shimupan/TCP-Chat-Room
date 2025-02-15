package main

import (
	"fmt"

	"github.com/shimupan/TCP-Chat-Room/pkg/server"
)

func main() {
	server := server.NewServer(":1337")
	err := server.Start()
	if err != nil {
		fmt.Printf("Err in server: %s\n", err)
	}

	return
}
