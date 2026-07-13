package ws

import (
	"encoding/json"
	"time"
)

type MessageType string

const (
	MsgTypeTaskStatus       MessageType = "task_status"
	MsgTypeNodeStatus       MessageType = "node_status"
	MsgTypeNodeLog          MessageType = "node_log"
	MsgTypeTaskDone         MessageType = "task_done"
	MsgTypeWorkflowProgress MessageType = "workflow_progress"
)

type Message struct {
	Type      MessageType `json:"type"`
	TaskID    string      `json:"taskId,omitempty"`
	NodeID    string      `json:"nodeId,omitempty"`
	Status    string      `json:"status,omitempty"`
	Progress  float64     `json:"progress,omitempty"`
	Payload   interface{} `json:"payload,omitempty"`
	Timestamp string      `json:"timestamp"`
}

type ClientMessage struct {
	Type string `json:"type"`
	Room string `json:"room"`
}

func NewMessage(msgType MessageType, payload interface{}) *Message {
	return &Message{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}
