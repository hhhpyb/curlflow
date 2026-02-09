package websocket

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ConnectionWrapper struct {
	Conn      *websocket.Conn
	Cancel    context.CancelFunc
	CreatedAt time.Time
}

type Manager struct {
	ctx         context.Context
	mu          sync.RWMutex
	connections map[string]*ConnectionWrapper
}

// NewManager creates a new WebSocket manager
func NewManager() *Manager {
	return &Manager{
		connections: make(map[string]*ConnectionWrapper),
	}
}

// SetContext sets the Wails runtime context
func (m *Manager) SetContext(ctx context.Context) {
	m.ctx = ctx
}

// WSMessage is the payload sent to frontend
type WSMessage struct {
	SessionID string `json:"session_id"`
	Type      string `json:"type"` // message, error, connected, disconnected
	Data      string `json:"data"`
}

// Connect establishes a new WebSocket connection for the given session ID
func (m *Manager) Connect(sessionID string, url string, headers map[string]string) error {
	// 1. Clean up existing session if any
	m.Disconnect(sessionID)

	// 2. Prepare Header
	header := http.Header{}
	for k, v := range headers {
		header.Add(k, v)
	}

	// 3. Connect
	dialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	conn, _, err := dialer.Dial(url, header)
	if err != nil {
		m.emit(sessionID, "error", fmt.Sprintf("Connection failed: %v", err))
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())

	wrapper := &ConnectionWrapper{
		Conn:      conn,
		Cancel:    cancel,
		CreatedAt: time.Now(),
	}

	m.mu.Lock()
	m.connections[sessionID] = wrapper
	m.mu.Unlock()

	m.emit(sessionID, "connected", "Connected to "+url)

	// 4. Start Read Loop
	go m.readLoop(ctx, sessionID, conn)

	return nil
}

// Disconnect closes the connection for the given session ID
func (m *Manager) Disconnect(sessionID string) {
	m.mu.Lock()
	wrapper, exists := m.connections[sessionID]
	if exists {
		delete(m.connections, sessionID)
	}
	m.mu.Unlock()

	if exists {
		wrapper.Cancel()
		wrapper.Conn.Close()
		m.emit(sessionID, "disconnected", "Disconnected")
	}
}

// SendMessage sends a text message to the specified session
func (m *Manager) SendMessage(sessionID string, message string) error {
	m.mu.RLock()
	wrapper, exists := m.connections[sessionID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("session not found")
	}

	err := wrapper.Conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		m.emit(sessionID, "error", fmt.Sprintf("Send failed: %v", err))
		return err
	}
	return nil
}

// CloseAll closes all active connections
func (m *Manager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, wrapper := range m.connections {
		wrapper.Cancel()
		wrapper.Conn.Close()
	}
	// Re-initialize map to be safe, though app is likely closing
	m.connections = make(map[string]*ConnectionWrapper)
}

// BatchMessage holds a batch of WebSocket messages
type BatchMessage struct {
	SessionID string      `json:"session_id"`
	Messages  []WSMessage `json:"messages"`
}

func (m *Manager) readLoop(ctx context.Context, sessionID string, conn *websocket.Conn) {
	defer func() {
		m.Disconnect(sessionID)
	}()

	// Channels for coordinating blocking reads with the ticker
	msgChan := make(chan string)
	errChan := make(chan error)

	// Start a goroutine for blocking reads
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				select {
				case errChan <- err:
				case <-ctx.Done():
				}
				return
			}
			select {
			case msgChan <- string(message):
			case <-ctx.Done():
				return
			}
		}
	}()

	buffer := make([]WSMessage, 0, 50)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	flush := func() {
		if len(buffer) == 0 {
			return
		}
		// Create a copy to safely emit
		msgsCopy := make([]WSMessage, len(buffer))
		copy(msgsCopy, buffer)

		payload := BatchMessage{
			SessionID: sessionID,
			Messages:  msgsCopy,
		}

		if m.ctx != nil {
			runtime.EventsEmit(m.ctx, "ws:batch-message", payload)
		}
		buffer = buffer[:0]
	}

	for {
		select {
		case <-ctx.Done():
			return
		case err := <-errChan:
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				m.emit(sessionID, "error", fmt.Sprintf("Read error: %v", err))
			}
			return
		case msg := <-msgChan:
			buffer = append(buffer, WSMessage{
				SessionID: sessionID,
				Type:      "message",
				Data:      msg,
			})
			if len(buffer) >= 50 {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func (m *Manager) emit(sessionID, msgType, data string) {
	if m.ctx == nil {
		return
	}
	payload := WSMessage{
		SessionID: sessionID,
		Type:      msgType,
		Data:      data,
	}
	runtime.EventsEmit(m.ctx, "ws:event", payload)
}
