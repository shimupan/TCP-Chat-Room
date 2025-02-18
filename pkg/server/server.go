package server

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/shimupan/TCP-Chat-Room/pkg/client"
	"github.com/shimupan/TCP-Chat-Room/pkg/helper"
)

type Server struct {
	addr        string
	listener    net.Listener
	mx          sync.RWMutex
	connections map[*client.Client]struct{}
	rooms       map[string][]*client.Client
	quitch      chan struct{}
}

func NewServer(addr string) *Server {
	return &Server{
		addr:        addr,
		mx:          sync.RWMutex{},
		connections: make(map[*client.Client]struct{}),
		rooms:       make(map[string][]*client.Client),
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
	helper.Logf("Server started listening on: %s\n", s.listener.Addr().String())

	go s.acceptConn()
	time.Sleep(1 * time.Millisecond)
	go s.handleCommands()

	<-s.quitch

	return nil
}

func (s *Server) acceptConn() {
	helper.Logf("Server accepting connections\n")
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quitch:
				helper.Logf("Server gracefully shut down, closing accept loop...\n")
				return
			default:
				helper.Logf("Error accept: %s\n", err)
				continue
			}
		}
		helper.Logf("New connection from %s\n", conn.RemoteAddr())
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	conn.Write([]byte("Please Enter Your Username: (at least 1 character long)\n"))
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		helper.Logf("Error reading username: %s\n", err)
		conn.Write([]byte("Error reading username, terminating session\n"))
		return
	}
	username := string(buf[:n])
	client := client.NewClient(username, conn)

	s.mx.Lock()
	s.connections[client] = struct{}{}
	s.mx.Unlock()

	conn.Write([]byte(fmt.Sprintf("Welcome %s!\n", buf[:n])))
	helper.Logf("Client %s has logged in as %s!\n", conn.RemoteAddr().String(), username)

	client.HandleCommands(s.rooms, &s.mx)
}

func (s *Server) handleCommands() {
	buf := make([]byte, 100)
	for {
		time.Sleep(5 * time.Millisecond)
		n, err := os.Stdin.Read(buf)
		if err != nil {
			helper.Logf("Error reading from stdin: %s\n", err)
			continue
		}
		command := strings.Split(strings.TrimSpace(string(buf[:n])), " ")
		switch command[0] {
		case "stop":
			s.stop()
			return
		case "list-users":
			s.listUsers()
		case "list-rooms":
			s.listRooms()
		case "kick":
			s.kickUser(command[1])
		}
	}
}

func (s *Server) stop() {
	helper.Logf("Server recieved stop command, gracefully shutting down...\n")
	close(s.quitch)
}

func (s *Server) listUsers() {
	if len(s.connections) == 0 {
		helper.Logf("There are no logged in clients currently...\n")
		return
	}
	cnt := 1
	for client := range s.connections {
		helper.Logf("%d) %s\n", cnt, client.ToString())
		cnt += 1
	}
}

func (s *Server) listRooms() {
	if len(s.rooms) == 0 {
		helper.Logf("There are no rooms currently...\n")
		return
	}

	for room := range s.rooms {
		fmt.Printf("1) room %v: ", room)
		for client := range s.rooms[room] {
			if client != len(s.rooms[room]) {
				fmt.Printf("[%s],", s.rooms[room][client].ToString())
			} else {
				fmt.Printf("[%s]", s.rooms[room][client].ToString())
			}
		}
	}
	helper.Logf("\n")
}

func (s *Server) kickUser(clientName string) {
	if len(s.connections) == 0 {
		helper.Logf("There are no logged in clients currently...\n")
		return
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	// Find and remove matching clients
	for client := range s.connections {
		if client.Username == clientName {
			err := client.Conn.Close()
			if err != nil {
				helper.Logf("Error kicking %s\n", client.ToString())
				continue
			}
			delete(s.connections, client)
			helper.Logf("Successfully kicked %s off the server\n", client.ToString())
		}
	}
}
