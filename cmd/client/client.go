package main

import (
	"fmt"
	"net"
)

func main() {
	_, err := net.Dial("tcp", ":1337")
	if err != nil {
		fmt.Printf("Error connecting to server: %s\n", err)
		return
	}
	fmt.Printf("Successfully connected to server!\n")
}
