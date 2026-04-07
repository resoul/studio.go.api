package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	ory "github.com/ory/client-go"
	"github.com/resoul/studio.go.api/internal/domain"
	"github.com/resoul/studio.go.api/internal/transport/http/utils"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Origin is checked by middleware, or add specific CORS origins here
	},
}

type WSHandler struct {
	hub            domain.PresenceHub
	profileService domain.ProfileService
}

func NewWSHandler(hub domain.PresenceHub, profileService domain.ProfileService) *WSHandler {
	return &WSHandler{
		hub:            hub,
		profileService: profileService,
	}
}

func (h *WSHandler) HandleWS(c *gin.Context) {
	user, ok := c.Get("user")
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "SNAKE_CASE_UNAUTHORIZED", "User not found in context")
		return
	}

	identity := user.(*ory.Identity)
	userID := identity.Id

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.Errorf("failed to upgrade to websocket: %v", err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan interface{}, 256),
	}

	h.hub.Register(context.Background(), userID, client)

	go h.writePump(client)
	go h.readPump(userID, client)
}

func (h *WSHandler) readPump(userID string, client *Client) {
	defer func() {
		h.hub.Unregister(context.Background(), userID, client)
		if err := h.profileService.MarkLastSeen(context.Background(), userID); err != nil {
			logrus.WithError(err).WithField("user_id", userID).Warn("failed to mark last seen")
		}
		client.conn.Close()
	}()

	client.conn.SetReadLimit(512)
	client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Errorf("websocket error: %v", err)
			}
			break
		}
	}
}

func (h *WSHandler) writePump(client *Client) {
	ticker := time.NewTicker(50 * time.Second)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.conn.WriteJSON(message); err != nil {
				return
			}
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Client helper struct for the handler, to match the one in infrastructure/ws
type Client struct {
	conn *websocket.Conn
	send chan interface{}
}

func (c *Client) Send(event interface{}) error {
	c.send <- event
	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
