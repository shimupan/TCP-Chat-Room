package client

import (
	"fmt"
	"io"
	"net"
)

type Client struct {
	username string
	conn     net.Conn
}

func NewClient(username string, conn net.Conn) *Client {
	return &Client{
		username: username,
		conn:     conn,
	}
}

func (c *Client) HandleCommands() {
	buf := make([]byte, 2048)
	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Client %s disconnected\n", c.conn.RemoteAddr().String())
				return
			}
			fmt.Printf("error reading: %s\n", err)
			continue
		}

		msg := buf[:n]
		fmt.Printf("Recieved msg: %s\n", msg)
	}
}
