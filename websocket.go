package base

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait       = 10 * time.Second    // Time allowed to write a message to the peer.
	pongWait        = 60 * time.Second    // Time allowed to read the next pong message from the peer.
	pingPeriod      = (pongWait * 9) / 10 // Send pings to peer with this period. Must be less than pongWait.
	maxMessageSize  = 512                 // Maximum message size allowed from peer.
	tokenValidation = int64(60 * time.Minute)
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { //该函数用于拦截或放行跨域请求。返回true放行，false拦截。
		return true
	},
}

var WebsocketHub *SocketHub

type SocketHub struct {
	clients map[*Client]bool // Registered clients.
	tokens  map[string]int64

	register     chan *Client //Register requests from the clients.
	unregister   chan *Client //Unregister requests from clients.
	tLock, cLock sync.Mutex   //lock for map
}

func NewHub() *SocketHub {
	WebsocketHub = &SocketHub{
		clients:    make(map[*Client]bool),
		tokens:     make(map[string]int64),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	return WebsocketHub
}

func (h *SocketHub) Run() {
	AccessLogger.Println("*SocketHub.Run")
	for {
		select {
		case client := <-h.register:
			h.cLock.Lock()
			h.clients[client] = true
			h.cLock.Unlock()
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.cLock.Lock()
				delete(h.clients, client)
				h.cLock.Unlock()
				if _, ok = h.tokens[client.token]; ok {
					h.tLock.Lock()
					delete(h.tokens, client.token)
					h.tLock.Unlock()
				}
				close(client.sendout)
			}
		}
		for k, v := range h.tokens {
			if v+tokenValidation < time.Now().UnixNano() {
				h.tLock.Lock()
				delete(h.tokens, k) //clean expired token
				h.tLock.Unlock()
			}
		}
	}
}

func (h *SocketHub) NewToken(token string, tm time.Time) {
	h.tLock.Lock()
	h.tokens[token] = tm.UnixNano()
	h.tLock.Unlock()
	return
}

func (h *SocketHub) NotifybyToken(token, msg string) {
	if _, ok := h.tokens[token]; ok {
		for c, _ := range h.clients {
			if c.token == token {
				c.Notify(msg)
				break
			}
		}
	}
}

type Client struct { // Client is a middleman between the websocket connection and the hub.
	token   string
	conn    *websocket.Conn // The websocket connection.
	sendout chan []byte     // Buffered channel of outbound messages.
}

func (c *Client) Notify(msg string) {
	c.sendout <- []byte(msg)
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		WebsocketHub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				ErrorLogger.Println("error:", err)
				AccessLogger.Println(err.Error())
				//websocket: close 1005 (no status) 客户端主动关闭
			}
			break
		}
		msg := string(message)
		AccessLogger.Println("recv:", msg)
	}
}

func (c *Client) readCommand(cmd chan string) {
	defer func() {
		WebsocketHub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				ss := err.Error()
				if !strings.Contains(ss, "websocket: close 1005 (no status)") {
					ErrorLogger.Println("error:", ss)
					AccessLogger.Println(ss) //The client actively closed.
				}
			}
			break
		}
		command := string(message)
		AccessLogger.Println("recv:", command)
		switch command {
		case "stop":
			cmd <- c.token + ":" + command
		case "background":
			cmd <- c.token + ":" + command
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.sendout:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				ErrorLogger.Println("c.conn.NextWriter:", err)
				return
			}
			_, err = w.Write(message)
			/*if err == nil {
				// Add queued chat messages to the current websocket message.
				n := len(c.sendout)
				for i := 0; i < n; i++ {
					//w.Write(newline)
					w.Write(<-c.sendout)
				}
			}*/

			if err = w.Close(); err != nil {
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

func WS_PicturetransferHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	token := r.FormValue("token")
	if _, ok := WebsocketHub.tokens[token]; ok {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			ErrorLogger.Println(err)
			return
		}
		AccessLogger.Println("register websocket hub -> Picturetransfer token:", token)
		client := &Client{token: token, conn: conn, sendout: make(chan []byte, 256)}
		WebsocketHub.register <- client

		go client.writePump()
		go client.readPump()
	}
}

func WS_DataexportHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	token := r.FormValue("token")
	if len(token) > 0 {
		if _, ok := WebsocketHub.tokens[token]; !ok {
			WebsocketHub.NewToken(token, time.Now())
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			ErrorLogger.Println("upgrader.Upgrade:", err.Error())
			AccessLogger.Println("upgrader.Upgrade:", err.Error())
			return
		}
		AccessLogger.Println("register websocket hub -> Dataexport token:", token)
		client := &Client{token: token, conn: conn, sendout: make(chan []byte, 256)}
		WebsocketHub.register <- client

		go client.writePump()
		go client.readCommand(DataexportCommandChannel)
	}
}
