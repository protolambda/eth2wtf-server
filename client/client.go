package client

import (
	"eth2wtf-server/common"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ClientHandler interface {
	Close()
	OnMessage(msg []byte)
	IsViewing(id common.ChunkID) bool
	common.ReceivePort
}

type MakeClientHandler func(send chan<- []byte) ClientHandler

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	// Maximum amounts of messages to buffer to a client before disconnecting them
	buffedMsgCount = 2000
)

var newline = []byte{'\n'}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	unregister func()

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	handler ClientHandler

	closed    bool
	closeLock sync.Mutex
}

func NewClient(conn *websocket.Conn, unregister func(), mkHandler MakeClientHandler) *Client {
	sendCh := make(chan []byte, buffedMsgCount)
	return &Client{
		unregister: unregister,
		conn:       conn,
		send:       sendCh,
		handler:    mkHandler(sendCh),
		closed:     false,
		closeLock:  sync.Mutex{},
	}
}

func (c *Client) IsViewing(id common.ChunkID) bool {
	return c.handler.IsViewing(id)
}

func (c *Client) Send(msg []byte) {
	c.handler.Send(msg)
}

func (c *Client) Close() {
	c.closeLock.Lock()
	c.handler.Close()
	c.closed = true
	close(c.send)
	c.closeLock.Unlock()
}

// ReadPump pumps messages from the websocket connection to the client message handler.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadPump() {
	defer func() {
		c.unregister()
		if err := c.conn.Close(); err != nil {
			fmt.Printf("Client %v unregistered with an error: %v\n", c, err)
		}
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.handler.OnMessage(message)
	}
	fmt.Println("quiting client")
}

// WritePump pumps messages from the client send channel to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err := c.conn.Close(); err != nil {
			fmt.Printf("Stopped connection with client %v, but with an error: %v\n", c, err)
		}
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			//fmt.Printf("%x\n", message)
			if _, err := w.Write(message); err != nil {
				fmt.Printf("Error when sending msg to client: %d, err: %v\n", c.id, err)
			}

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendMsg(msg []byte) {
	c.closeLock.Lock()
	if !c.closed {
		c.send <- msg
	}
	c.closeLock.Unlock()
}
