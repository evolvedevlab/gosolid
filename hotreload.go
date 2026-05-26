package gowebi

import (
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const browserReloadScript = `
<script>
const ws = new WebSocket("ws://localhost:6969/ws");
ws.onopen = () => console.log("dev websocket connected");
ws.onmessage = (e) => {
  if (e.data === "reload") {
    location.reload();
  }
};
ws.onclose = () => {
  console.log("dev websocket disconnected, retrying...");
  setTimeout(() => {
  	setTimeout(() => {
        window.location.reload();
    }, 500);
  }, 300);
};
ws.onerror = () => {
  ws.close();
};
</script>`

type wsHandler struct {
	watchPath string
	upg       *websocket.Upgrader
	conns     map[*websocket.Conn]bool
	mu        sync.RWMutex
}

func newWSHandler(watchPath string) *wsHandler {
	return &wsHandler{
		watchPath: watchPath,
		conns:     make(map[*websocket.Conn]bool),
		upg: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (ws *wsHandler) start() error {
	go ws.loop()

	http.HandleFunc("/ws", ws.handleWS)
	return http.ListenAndServe(":6969", nil)
}

func (ws *wsHandler) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upg.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	ws.mu.Lock()
	ws.conns[conn] = true
	ws.mu.Unlock()
}

func (ws *wsHandler) loop() {
	ticker := time.NewTicker(time.Millisecond * 500)
	defer ticker.Stop()

	var lastMod time.Time
	for range ticker.C {
		info, err := os.Stat(ws.watchPath)
		if err != nil {
			continue
		}

		mod := info.ModTime()
		if mod.After(lastMod) {
			ws.broadcast()

			lastMod = mod
		}
	}
}

func (ws *wsHandler) broadcast() {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	for conn := range ws.conns {
		conn.WriteMessage(websocket.TextMessage, []byte("reload"))
	}
}

func (ws *wsHandler) removeConn(conn *websocket.Conn) error {
	ws.mu.Lock()
	delete(ws.conns, conn)
	ws.mu.Unlock()

	return conn.Close()
}
