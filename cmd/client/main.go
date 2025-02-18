package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", ":1337")
	if err != nil {
		fmt.Printf("Error connecting to server: %s\n", err)
		return
	}
	defer conn.Close()
	fmt.Printf("Successfully connected to server!\n")

	clientch := make(chan struct{})

	go listener(conn, clientch)
	fmt.Printf("Successfully setup listener!\n")
	go writer(conn, clientch)
	fmt.Printf("Successfully setup writer!\n")

	<-clientch
}

func listener(conn net.Conn, clientch chan struct{}) {
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Listener: Server closed connection...\n")
				close(clientch)
				return
			}
			fmt.Printf("Listener: Error reading from server: %s\n", err)
			continue
		}
		msg := string(buf[:n])
		fmt.Printf("Msg from server: %s", msg)
	}
}

func writer(conn net.Conn, clientch chan struct{}) {
	buf := make([]byte, 2048)
	for {
		// Read command from stdin
		n, err := os.Stdin.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Writer: Server closed connection...: %s\n", err)
				return
			}
			fmt.Printf("Writer: Error reading from stdin: %s\n", err)
			return
		}

		// Write command to server
		command := strings.TrimRight(string(buf[:n]), "\n")
		_, err = conn.Write([]byte(command))
		if err != nil {
			fmt.Printf("Writer: Error writing to server, connection might be closed: %s\n", err)
			close(clientch)
			return
		}
		fmt.Printf("Successfully wrote %s to server!\n", command)
	}
}
