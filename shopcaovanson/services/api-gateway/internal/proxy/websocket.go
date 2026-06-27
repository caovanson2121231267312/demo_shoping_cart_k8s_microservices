package proxy

import (
	"net/url"
	"strings"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

var wsUpgrader = websocket.FastHTTPUpgrader{
	CheckOrigin: func(_ *fasthttp.RequestCtx) bool {
		return true
	},
}

func WebSocketHandler(baseURL string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !websocket.FastHTTPIsWebSocketUpgrade(c.Context()) {
			return fiber.ErrUpgradeRequired
		}

		targetURL, err := buildWebSocketTarget(baseURL, c.OriginalURL())
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error": "invalid websocket target",
			})
		}

		return wsUpgrader.Upgrade(c.Context(), func(clientConn *websocket.Conn) {
			defer clientConn.Close()

			remoteConn, _, err := websocket.DefaultDialer.Dial(targetURL, nil)
			if err != nil {
				return
			}
			defer remoteConn.Close()

			errc := make(chan error, 2)
			go pumpWebSocket(clientConn, remoteConn, errc)
			go pumpWebSocket(remoteConn, clientConn, errc)
			<-errc
		})
	}
}

func buildWebSocketTarget(baseURL, originalURL string) (string, error) {
	base := strings.TrimRight(baseURL, "/")
	target := base + originalURL
	parsed, err := url.Parse(target)
	if err != nil {
		return "", err
	}
	switch parsed.Scheme {
	case "http":
		parsed.Scheme = "ws"
	case "https":
		parsed.Scheme = "wss"
	}
	return parsed.String(), nil
}

func pumpWebSocket(dst, src *websocket.Conn, errc chan<- error) {
	for {
		messageType, payload, err := src.ReadMessage()
		if err != nil {
			errc <- err
			return
		}
		if err := dst.WriteMessage(messageType, payload); err != nil {
			errc <- err
			return
		}
	}
}
