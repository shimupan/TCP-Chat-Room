package server

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/shimupan/TCP-Chat-Room/pkg/client"
)

type Server struct {
	addr        string
	listener    net.Listener
	mx          sync.RWMutex
	connections map[*client.Client]struct{}
	quitch      chan struct{}
}

func NewServer(addr string) *Server {
	return &Server{
		addr:        addr,
		mx:          sync.RWMutex{},
		connections: make(map[*client.Client]struct{}),
		quitch:      make(chan struct{}),
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer listener.Close()
	s.listener = listener
	fmt.Printf("Server started listening on: %s\n", s.listener.Addr().String())

	go s.acceptConn()
	go s.handleCommands()

	<-s.quitch

	return nil
}

func (s *Server) acceptConn() {
	fmt.Printf("Server accepting connections\n")
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quitch:
				fmt.Printf("Server gracefully shut down, closing accept loop...\n")
				return
			default:
				fmt.Printf("Error accept: %s\n", err)
				continue
			}
		}
		fmt.Printf("New connection from %s\n", conn.RemoteAddr())
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	conn.Write([]byte("Please Enter Your Username: (at least 1 character long)\n"))
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("Error reading username: %s\n", err)
		conn.Write([]byte("Error reading username, terminating session\n"))
		return
	}
	username := string(buf[:n])
	client := client.NewClient(username, conn)

	s.mx.RLock()
	s.connections[client] = struct{}{}
	s.mx.RUnlock()

	conn.Write([]byte(fmt.Sprintf("Welcome %s!\n", buf[:n])))
	fmt.Printf("Client %s has logged in as %s!\n", conn.RemoteAddr().String(), username)

	client.HandleCommands()
}

func (s *Server) handleCommands() {
	buf := make([]byte, 100)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			fmt.Printf("Error reading from stdin: %s\n", err)
			continue
		}
		command := strings.TrimSpace(string(buf[:n]))
		switch command {
		case "stop":
			fmt.Printf("Server recieved stop command, gracefully shutting down...\n")
			close(s.quitch)
			return
		}
	}
}
