package handlers

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aistudio/backend/internal/service"
	"github.com/aistudio/backend/internal/task"
)

// WebSocketEvent maps to the frontend WebSocketEvent interface.
type WebSocketEvent struct {
	Type   string            `json:"type"`
	TaskID string            `json:"taskId"`
	Data   WebSocketEventData `json:"data"`
}

// WebSocketEventData contains the payload of a WebSocket event.
type WebSocketEventData struct {
	Status    string      `json:"status,omitempty"`
	Progress  float64     `json:"progress,omitempty"`
	Message   string      `json:"message,omitempty"`
	Level     string      `json:"level,omitempty"`
	Step      string      `json:"step,omitempty"`
	Error     string      `json:"error,omitempty"`
	Result    interface{} `json:"result,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// wsConn wraps a net.Conn with WebSocket frame handling.
type wsConn struct {
	conn   net.Conn
	reader *bufio.Reader
	mu     sync.Mutex
}

// WebSocketHandler manages WebSocket connections and broadcasts task events.
type WebSocketHandler struct {
	clients  map[*wsConn]bool
	mu       sync.RWMutex
	eventBus *task.EventBus
	done     chan struct{}
}

// NewWebSocketHandler creates a new WebSocket handler.
func NewWebSocketHandler(svc *service.TaskService) *WebSocketHandler {
	h := &WebSocketHandler{
		clients:  make(map[*wsConn]bool),
		eventBus: svc.EventBus(),
		done:     make(chan struct{}),
	}

	// Subscribe to all task events and broadcast to connected clients
	h.eventBus.SubscribeAll(func(event *task.TaskEvent) {
		h.broadcast(event)
	})

	return h
}

// HandleWebSocket handles the WebSocket upgrade and connection lifecycle.
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Perform WebSocket upgrade handshake
	conn, err := upgrade(w, r)
	if err != nil {
		log.Printf("[ws] upgrade failed: %v", err)
		http.Error(w, "WebSocket upgrade failed", http.StatusBadRequest)
		return
	}

	ws := &wsConn{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}

	h.mu.Lock()
	h.clients[ws] = true
	h.mu.Unlock()

	log.Printf("[ws] client connected (total: %d)", len(h.clients))

	// Read loop: keep connection alive and handle close
	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.clients, ws)
			h.mu.Unlock()
			ws.conn.Close()
			log.Printf("[ws] client disconnected (total: %d)", len(h.clients))
		}()

		for {
			opcode, payload, err := readFrame(ws)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					log.Printf("[ws] read error: %v", err)
				}
				return
			}

			switch opcode {
			case opPing:
				// Respond with pong
				writeFrame(ws, opPong, payload)
			case opClose:
				writeFrame(ws, opClose, nil)
				return
			case opPong:
				// Client responded to our ping, nothing to do
			case opText:
				// Client messages are ignored for now
			}
		}
	}()

	// Ping ticker to keep connection alive
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				h.mu.RLock()
				_, exists := h.clients[ws]
				h.mu.RUnlock()
				if !exists {
					return
				}
				if err := writeFrame(ws, opPing, []byte{}); err != nil {
					return
				}
			case <-h.done:
				return
			}
		}
	}()
}

// broadcast sends a task event to all connected WebSocket clients.
func (h *WebSocketHandler) broadcast(event *task.TaskEvent) {
	wsEvent := h.mapTaskEvent(event)
	data, err := json.Marshal(wsEvent)
	if err != nil {
		log.Printf("[ws] marshal error: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for ws := range h.clients {
		if err := writeFrame(ws, opText, data); err != nil {
			log.Printf("[ws] write error: %v", err)
			ws.conn.Close()
			go func(c *wsConn) {
				h.mu.Lock()
				delete(h.clients, c)
				h.mu.Unlock()
			}(ws)
		}
	}
}

// Close stops the ping ticker and closes all connections.
func (h *WebSocketHandler) Close() {
	close(h.done)
	h.mu.Lock()
	defer h.mu.Unlock()
	for ws := range h.clients {
		ws.conn.Close()
	}
	h.clients = make(map[*wsConn]bool)
}

// mapTaskEvent converts a backend TaskEvent to a frontend WebSocketEvent.
func (h *WebSocketHandler) mapTaskEvent(event *task.TaskEvent) WebSocketEvent {
	ts := event.Timestamp.Format(time.RFC3339)

	switch event.Type {
	case task.EventTaskCreated:
		return WebSocketEvent{
			Type:   "task_status",
			TaskID: event.TaskID,
			Data: WebSocketEventData{
				Status:    string(event.Status),
				Timestamp: ts,
			},
		}
	case task.EventTaskStarted:
		return WebSocketEvent{
			Type:   "task_status",
			TaskID: event.TaskID,
			Data: WebSocketEventData{
				Status:    string(event.Status),
				Progress:  event.Progress,
				Timestamp: ts,
			},
		}
	case task.EventTaskProgress:
		msg := ""
		if event.Data != nil {
			if s, ok := event.Data.(string); ok {
				msg = s
			} else if b, err := json.Marshal(event.Data); err == nil {
				msg = string(b)
			}
		}
		return WebSocketEvent{
			Type:   "task_progress",
			TaskID: event.TaskID,
			Data: WebSocketEventData{
				Status:    string(event.Status),
				Progress:  event.Progress,
				Message:   msg,
				Timestamp: ts,
			},
		}
	case task.EventTaskCompleted:
		return WebSocketEvent{
			Type:   "task_complete",
			TaskID: event.TaskID,
			Data: WebSocketEventData{
				Status:    string(event.Status),
				Progress:  100,
				Result:    event.Data,
				Timestamp: ts,
			},
		}
	case task.EventTaskFailed:
		errMsg := ""
		if event.Data != nil {
			if s, ok := event.Data.(string); ok {
				errMsg = s
			}
		}
		return WebSocketEvent{
			Type:   "task_error",
			TaskID: event.TaskID,
			Data: WebSocketEventData{
				Status:    string(event.Status),
				Error:     errMsg,
				Timestamp: ts,
			},
		}
	case task.EventTaskCancelled:
		return WebSocketEvent{
			Type:   "task_error",
			TaskID: event.TaskID,
			Data: WebSocketEventData{
				Status:    string(event.Status),
				Error:     "task cancelled",
				Timestamp: ts,
			},
		}
	default:
		return WebSocketEvent{
			Type:   "task_status",
			TaskID: event.TaskID,
			Data: WebSocketEventData{
				Status:    string(event.Status),
				Timestamp: ts,
			},
		}
	}
}

// ---- WebSocket protocol implementation (RFC 6455) ----

const (
	// WebSocket opcodes
	opCont  = 0x0
	opText  = 0x1
	opBin   = 0x2
	opClose = 0x8
	opPing  = 0x9
	opPong  = 0xA

	// WebSocket GUID for handshake
	wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
)

// upgrade performs the WebSocket handshake and returns the raw net.Conn.
func upgrade(w http.ResponseWriter, r *http.Request) (net.Conn, error) {
	if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		return nil, errors.New("missing Upgrade: websocket header")
	}
	if !headerContains(r.Header, "Connection", "upgrade") {
		return nil, errors.New("missing Connection: upgrade header")
	}

	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		return nil, errors.New("missing Sec-WebSocket-Key header")
	}

	// Compute accept key
	h := sha1.New()
	h.Write([]byte(key + wsGUID))
	acceptKey := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Hijack the connection
	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, errors.New("server does not support hijacking")
	}

	conn, bufrw, err := hj.Hijack()
	if err != nil {
		return nil, err
	}

	// Write the 101 Switching Protocols response
	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + acceptKey + "\r\n\r\n"

	if _, err := bufrw.WriteString(resp); err != nil {
		conn.Close()
		return nil, err
	}
	if err := bufrw.Flush(); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

// readFrame reads a single WebSocket frame from the connection.
func readFrame(ws *wsConn) (opcode byte, payload []byte, err error) {
	// Read the first 2 bytes (header)
	header := make([]byte, 2)
	if _, err := io.ReadFull(ws.reader, header); err != nil {
		return 0, nil, err
	}

	opcode = header[0] & 0x0F
	masked := header[1]&0x80 != 0
	length := uint64(header[1] & 0x7F)

	// Extended payload length
	switch length {
	case 126:
		ext := make([]byte, 2)
		if _, err := io.ReadFull(ws.reader, ext); err != nil {
			return 0, nil, err
		}
		length = uint64(binary.BigEndian.Uint16(ext))
	case 127:
		ext := make([]byte, 8)
		if _, err := io.ReadFull(ws.reader, ext); err != nil {
			return 0, nil, err
		}
		length = binary.BigEndian.Uint64(ext)
	}

	// Read mask key
	var maskKey [4]byte
	if masked {
		if _, err := io.ReadFull(ws.reader, maskKey[:]); err != nil {
			return 0, nil, err
		}
	}

	// Read payload
	if length > 0 {
		payload = make([]byte, length)
		if _, err := io.ReadFull(ws.reader, payload); err != nil {
			return 0, nil, err
		}
		if masked {
			for i := range payload {
				payload[i] ^= maskKey[i%4]
			}
		}
	}

	return opcode, payload, nil
}

// writeFrame writes a single WebSocket frame to the connection.
func writeFrame(ws *wsConn, opcode byte, payload []byte) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Build frame header
	frame := make([]byte, 0, 10+len(payload))

	// FIN + opcode
	frame = append(frame, 0x80|opcode)

	// Mask + payload length (server frames are NOT masked per RFC 6455)
	length := len(payload)
	switch {
	case length <= 125:
		frame = append(frame, byte(length))
	case length <= 65535:
		frame = append(frame, 126)
		ext := make([]byte, 2)
		binary.BigEndian.PutUint16(ext, uint16(length))
		frame = append(frame, ext...)
	default:
		frame = append(frame, 127)
		ext := make([]byte, 8)
		binary.BigEndian.PutUint64(ext, uint64(length))
		frame = append(frame, ext...)
	}

	// Payload
	frame = append(frame, payload...)

	_, err := ws.conn.Write(frame)
	return err
}

// headerContains checks if the header value contains the given string (case-insensitive).
func headerContains(h http.Header, key, value string) bool {
	for _, v := range h[http.CanonicalHeaderKey(key)] {
		for _, part := range strings.Split(v, ",") {
			if strings.EqualFold(strings.TrimSpace(part), value) {
				return true
			}
		}
	}
	return false
}