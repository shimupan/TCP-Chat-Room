package client

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/shimupan/TCP-Chat-Room/pkg/helper"
)

type Client struct {
	Username string
	Conn     net.Conn
	Room     *helper.RoomInfo
}

func NewClient(username string, conn net.Conn) *Client {
	return &Client{
		Username: username,
		Conn:     conn,
		Room: &helper.RoomInfo{
			RoomId: "",
			Owner:  false,
		},
	}
}

func (c *Client) HandleCommands(rooms map[string][]*Client, mx *sync.RWMutex) {
	buf := make([]byte, 2048)
	for {
		n, err := c.Conn.Read(buf)
		if err != nil {
			if err == io.EOF || isConnectionClosed(err) {
				helper.Logf("Client %s disconnected\n", c.Conn.RemoteAddr().String())
				return
			}
			helper.Logf("error reading: %s\n", err)
			continue
		}

		msg := strings.Split(string(buf[:n]), " ")
		switch msg[0] {
		case "-create":
			c.createRoom(msg[1], rooms, mx)
		case "-delete":
			c.deleteRoom(msg[1], rooms, mx)
		}
		helper.Logf("Recieved msg: %s\n", msg)
	}
}

func (c *Client) createRoom(joining string, rooms map[string][]*Client, mx *sync.RWMutex) {
	if c.Room.RoomId != "" {
		c.Conn.Write([]byte(fmt.Sprintf("You are already in room %s, please leave before making another room\n", c.Room.RoomId)))
		return
	}
	mx.Lock()
	defer mx.Unlock()
	// Create new room if not exists
	if _, exists := rooms[joining]; !exists {
		rooms[joining] = make([]*Client, 0)
	}

	// Add client to room
	rooms[joining] = append(rooms[joining], c)
	c.Room.RoomId = joining
	c.Room.Owner = true

	c.Conn.Write([]byte(fmt.Sprintf("Successfully created/joined room: %s\n", joining)))
	helper.Logf("Client %s created and joined room %s\n", c.ToString(), joining)
}

func (c *Client) deleteRoom(deleting string, rooms map[string][]*Client, mx *sync.RWMutex) {
	if c.Room.RoomId == "" {
		c.Conn.Write([]byte("You are currently not in a room\n"))
		return
	} else if !c.Room.Owner {
		c.Conn.Write([]byte(fmt.Sprintf("You are not the owner of room %s\n", c.Room.RoomId)))
		return
	}
	mx.Lock()
	defer mx.Unlock()
	client_list := rooms[deleting]

	for cli := range client_list {
		client_list[cli].LeaveRoom()
	}
	delete(rooms, deleting)
}

func (c *Client) LeaveRoom() {
	c.Room.RoomId = ""
	c.Room.Owner = false
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
