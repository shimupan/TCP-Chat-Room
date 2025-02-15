package main

import (
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", ":1337")
	if err != nil {
		fmt.Printf("Error connecting to server: %s\n", err)
		return
	}
	defer conn.Close()

	fmt.Printf("Successfully connected to server!\n")

	n, err := conn.Write([]byte("Hello Server!"))
	if err != nil {
		fmt.Printf("Error writing to server: %s\n", err)
		return
	}

	fmt.Printf("Wrote %d bytes to server!\n", n)

	n, err = conn.Write([]byte(""))
	if err != nil {
		fmt.Printf("Error writing to server: %s\n", err)
		return
	}

	fmt.Printf("Wrote %d bytes to server!\n", n)
}
