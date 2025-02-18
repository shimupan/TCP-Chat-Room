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
	c.Conn.Write([]byte(c.listCommands()))
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
		case "-join":
			c.joinRoom(msg[1], rooms, mx)
		case "-leave":
			c.leaveRoom(c.Room.RoomId, rooms, mx)
		case "-me":
			c.me()
		case "-anyone":
			c.anyone(rooms)
		case "-help":
			c.Conn.Write([]byte(c.listCommands()))
		default:
			c.handleMessage(strings.Join(msg, " "), rooms)
		}
		// helper.Logf("Recieved msg: %s\n", msg)
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
	} else {
		c.Conn.Write([]byte(fmt.Sprintf("Room %s already exist, please use -join <room> to join it\n", c.Room.RoomId)))
		return
	}

	// Add client to room
	rooms[joining] = append(rooms[joining], c)
	c.Room.RoomId = joining
	c.Room.Owner = true

	c.Conn.Write([]byte(fmt.Sprintf("Successfully created/joined room: %s!\nYour messages will now be directed to the current room you're in!\n", joining)))
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
	helper.Logf("Client %s deleted and removed everyone from room %s\n", c.ToString(), deleting)
}

func (c *Client) joinRoom(joining string, rooms map[string][]*Client, mx *sync.RWMutex) {
	if c.Room.RoomId != "" {
		c.Conn.Write([]byte(fmt.Sprintf("You are already in a room: %s\n", c.Room.RoomId)))
		return
	}
	if _, exists := rooms[joining]; !exists {
		c.Conn.Write([]byte("This room does not exist, please create it if you want\n"))
		return
	}

	c.JoinRoom(joining)

	client_list := rooms[joining]
	for cli := range client_list {
		client_list[cli].Conn.Write([]byte(fmt.Sprintf("%s has joined the room!\n", c.Username)))
	}

	mx.Lock()
	rooms[joining] = append(rooms[joining], c)
	mx.Unlock()
}

func (c *Client) leaveRoom(leaving string, rooms map[string][]*Client, mx *sync.RWMutex) {
	if c.Room.RoomId == "" {
		c.Conn.Write([]byte("You are currently not in a room\n"))
		return
	}

	mx.Lock()
	defer mx.Unlock()
	client_list := rooms[leaving]
	client_list[0].LeaveRoom()

	if len(rooms[leaving]) == 1 {
		delete(rooms, leaving)
		rooms[leaving][0].Conn.Write([]byte(fmt.Sprintf("You have been left room %s\n", leaving)))
		helper.Logf("Client %s left the room %s and it got deleted\n", c.ToString(), leaving)
	} else {
		rooms[leaving] = client_list[1:]
		rooms[leaving][0].Room.Owner = true
		rooms[leaving][0].Conn.Write([]byte("You have been promoted to the leader of this room!\n"))
		helper.Logf("Client %s left the room %s and %s became leader!\n", c.ToString(), leaving, rooms[leaving][0].ToString())
	}
}

func (c *Client) me() {
	c.Conn.Write([]byte(fmt.Sprintf("You are: %s\n", c.ToString())))
	helper.Logf("Client %s requested information about themselves\n", c.ToString())
}

func (c *Client) anyone(rooms map[string][]*Client) {
	if c.Room.RoomId == "" {
		c.Conn.Write([]byte("You are not in any room\n"))
		return
	}

	curr_room := rooms[c.Room.RoomId]
	var members string
	for i, client := range curr_room {
		if i < len(curr_room)-1 {
			members += fmt.Sprintf("[%s], ", client.ToString())
		} else {
			members += fmt.Sprintf("[%s]", client.ToString())
		}
	}

	response := fmt.Sprintf("Room %s members:\n%s\n", c.Room.RoomId, members)
	c.Conn.Write([]byte(response))
	helper.Logf("Client %s requested information about their room\n", c.ToString())
}

func (c *Client) handleMessage(msg string, rooms map[string][]*Client) {
	if c.Room.RoomId == "" {
		c.Conn.Write([]byte("You must be in a room before being able to chat, please type [-help] if you need help\n"))
		helper.Logf("Client %s tried to message but wasn't in a room\n", c.ToString())
		return
	}

	client_list := rooms[c.Room.RoomId]

	for cli := range client_list {
		client_list[cli].Conn.Write([]byte(fmt.Sprintf("%s: %s\n", c.Username, msg)))
	}

	helper.Logf("Client %s sent a message to room %s\n", c.ToString(), c.Room.RoomId)
}

// Helper function to check if the error is due to closed connection
func isConnectionClosed(err error) bool {
	if opErr, ok := err.(*net.OpError); ok {
		return opErr.Err.Error() == "use of closed network connection"
	}
	return false
}

func (c *Client) listCommands() string {
	commands := []struct {
		cmd         string
		description string
	}{
		{"Hey there!", "Try some of these commands!"},
		{"-create <room>", "Create and join a new chat room"},
		{"-delete <room>", "Delete a room (must be owner and in a room)"},
		{"-join <room>", "Join a room (the room must already exist)"},
		{"-leave", "Leave your current room"},
		{"-me", "Show your current user information"},
		{"-anyone", "List all members in your current room"},
		{"-help", "Show this help message"},
	}

	var helpText strings.Builder
	helpText.WriteString("Available commands:\n")
	for _, cmd := range commands {
		helpText.WriteString(fmt.Sprintf("%-20s - %s\n", cmd.cmd, cmd.description))
	}

	return helpText.String()
}

func (c *Client) LeaveRoom() {
	c.Room.RoomId = ""
	c.Room.Owner = false
}

func (c *Client) JoinRoom(room string) {
	c.Room.RoomId = room
	c.Room.Owner = false
}

func (c *Client) ToString() string {
	return fmt.Sprintf("%s:%s, room:%s, owner:%t", c.Username, c.Conn.RemoteAddr().String(), c.Room.RoomId, c.Room.Owner)
}
