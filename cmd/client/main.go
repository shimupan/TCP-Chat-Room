package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
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
	conn.SetReadDeadline(time.Now().Add(2 * time.Second)) // set up non blocking read
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if os.IsTimeout(err) { // Read timed out, check if writer exited
				select {
				case <-clientch:
					fmt.Println("Writer exited, closing listener.")
					return
				default:
					continue // Keep listening
				}
			}
			fmt.Printf("Error reading from server: %s\n", err)
			continue
		}
		msg := string(buf[:n])
		fmt.Printf("Msg from server: %s", msg)
	}
}

func writer(conn net.Conn, clientch chan struct{}) {
	buf := make([]byte, 2048)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			if err == io.EOF {
				close(clientch)
				return
			}
			fmt.Printf("Error writing: %s\n", err)
			continue
		}

		command := strings.TrimRight(string(buf[:n]), "\n")
		_, err = conn.Write([]byte(command))
		if err != nil {
			fmt.Printf("Error writing to server: %s\n", err)
			continue
		}
		fmt.Printf("Successfully wrote %s to server!\n", command)
	}
}
