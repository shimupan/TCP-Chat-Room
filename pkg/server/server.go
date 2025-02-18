package server

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

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
	s.logf("Server started listening on: %s\n", s.listener.Addr().String())

	go s.acceptConn()
	time.Sleep(1 * time.Millisecond)
	go s.handleCommands()

	<-s.quitch

	return nil
}

func (s *Server) acceptConn() {
	s.logf("Server accepting connections\n")
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quitch:
				s.logf("Server gracefully shut down, closing accept loop...\n")
				return
			default:
				s.logf("Error accept: %s\n", err)
				continue
			}
		}
		s.logf("New connection from %s\n", conn.RemoteAddr())
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	conn.Write([]byte("Please Enter Your Username: (at least 1 character long)\n"))
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		s.logf("Error reading username: %s\n", err)
		conn.Write([]byte("Error reading username, terminating session\n"))
		return
	}
	username := string(buf[:n])
	client := client.NewClient(username, conn)

	s.mx.RLock()
	s.connections[client] = struct{}{}
	s.mx.RUnlock()

	conn.Write([]byte(fmt.Sprintf("Welcome %s!\n", buf[:n])))
	s.logf("Client %s has logged in as %s!\n", conn.RemoteAddr().String(), username)

	client.HandleCommands()
}

func (s *Server) handleCommands() {
	buf := make([]byte, 100)
	for {
		time.Sleep(5 * time.Millisecond)
		n, err := os.Stdin.Read(buf)
		if err != nil {
			s.logf("Error reading from stdin: %s\n", err)
			continue
		}
		command := strings.Split(strings.TrimSpace(string(buf[:n])), " ")
		switch command[0] {
		case "stop":
			s.stop()
			return
		case "list":
			s.listUsers()
		case "kick":
			s.kickUser(command[1])
		}
	}
}

func (s *Server) stop() {
	s.logf("Server recieved stop command, gracefully shutting down...\n")
	close(s.quitch)
}

func (s *Server) listUsers() {
	if len(s.connections) == 0 {
		s.logf("There are no logged in clients currently...\n")
		return
	}
	cnt := 1
	for client := range s.connections {
		s.logf("%d) %s\n", cnt, client.ToString())
		cnt += 1
	}
}

func (s *Server) kickUser(clientName string) {
	if len(s.connections) == 0 {
		s.logf("There are no logged in clients currently...\n")
		return
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	// Find and remove matching clients
	for client := range s.connections {
		if client.Username == clientName {
			err := client.Conn.Close()
			if err != nil {
				s.logf("Error kicking %s\n", client.ToString())
				continue
			}
			delete(s.connections, client)
			s.logf("Successfully kicked %s off the server\n", client.ToString())
		}
	}
}

func (s *Server) logf(format string, args ...interface{}) {
	fmt.Printf("\r"+format+"> ", args...)
}
