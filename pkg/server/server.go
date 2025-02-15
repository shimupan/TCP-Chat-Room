package server

import (
	"fmt"
	"io"
	"net"
)

type Server struct {
	addr     string
	listener net.Listener
	quitch   chan struct{}
}

func NewServer(addr string) *Server {
	return &Server{
		addr:   addr,
		quitch: make(chan struct{}),
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

	<-s.quitch

	return nil
}

func (s *Server) acceptConn() {
	fmt.Printf("Server accepting connections\n")
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Printf("Error accept: %s\n", err)
			continue
		}
		fmt.Printf("New connection from %s\n", conn.RemoteAddr())
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Client %s disconnected\n", conn.RemoteAddr().String())
				return
			}
			fmt.Printf("error reading: %s\n", err)
			continue
		}

		msg := buf[:n]
		fmt.Printf("Recieved msg: %s\n", msg)
	}
}
