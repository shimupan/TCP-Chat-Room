package client

import (
	"fmt"
	"io"
	"net"
)

type Client struct {
	Username string
	Conn     net.Conn
}

func NewClient(username string, conn net.Conn) *Client {
	return &Client{
		Username: username,
		Conn:     conn,
	}
}

func (c *Client) HandleCommands() {
	buf := make([]byte, 2048)
	for {
		n, err := c.Conn.Read(buf)
		if err != nil {
			if err == io.EOF || isConnectionClosed(err) {
				logf("Client %s disconnected\n", c.Conn.RemoteAddr().String())
				return
			}
			logf("error reading: %s\n", err)
			continue
		}

		msg := buf[:n]
		logf("Recieved msg: %s\n", msg)
	}
}

func (c *Client) ToString() string {
	return fmt.Sprintf("%s:%s", c.Username, c.Conn.RemoteAddr().String())
}

// Helper function to check if the error is due to closed connection
func isConnectionClosed(err error) bool {
	if opErr, ok := err.(*net.OpError); ok {
		return opErr.Err.Error() == "use of closed network connection"
	}
	return false
}

func logf(format string, args ...interface{}) {
	fmt.Printf("\r"+format+"> ", args...)
}
