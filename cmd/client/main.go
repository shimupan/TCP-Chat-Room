package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/shimupan/TCP-Chat-Room/pkg/helper"
)

func main() {
	conn, err := net.Dial("tcp", ":1337")
	if err != nil {
		helper.Logf("Error connecting to server: %s\n", err)
		return
	}
	defer conn.Close()
	helper.Logf("Successfully connected to server!\n")

	clientch := make(chan struct{})

	go listener(conn, clientch)
	helper.Logf("Successfully setup listener!\n")
	go writer(conn, clientch)
	helper.Logf("Successfully setup writer!\n")

	<-clientch
}

func listener(conn net.Conn, clientch chan struct{}) {
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				helper.Logf("Listener: Server closed connection...\n")
				close(clientch)
				return
			}
			helper.Logf("Listener: Error reading from server: %s\n", err)
			continue
		}
		msg := string(buf[:n])
		helper.Logf("Server: %s", msg)
	}
}

func writer(conn net.Conn, clientch chan struct{}) {
	buf := make([]byte, 2048)
	i := 0
	for {
		// Only show prompt for commands
		fmt.Printf("> ")

		// Read command from stdin
		n, err := os.Stdin.Read(buf)
		if err != nil {
			if err == io.EOF {
				helper.Logf("Writer: Server closed connection...: %s\n", err)
				return
			}
			helper.Logf("Writer: Error reading from stdin: %s\n", err)
			return
		}

		// Clear the prompt line
		if i != 0 {
			fmt.Print("\033[1A\033[K") // Move up one line and clear it
		}

		command := strings.TrimRight(string(buf[:n]), "\n")

		if strings.HasPrefix(command, "-") {
			helper.Logf("> %s\n", command)
		}

		_, err = conn.Write([]byte(command))
		if err != nil {
			helper.Logf("Writer: Error writing to server, connection might be closed: %s\n", err)
			close(clientch)
			return
		}
		i += 1
	}
}
