package agent

import (
	"sync"
	"time"

	"github.com/aistudio/packages/workflow"
)

type ConversationStore interface {
	Save(conversation *Conversation) error
	Get(id string) (*Conversation, error)
	List(limit, offset int) ([]*Conversation, error)
	Delete(id string) error
}

type Memory struct {
	mu           sync.RWMutex
	conversations map[string]*Conversation
	maxHistory   int
	store        ConversationStore
}

func NewMemory(maxHistory int) *Memory {
	if maxHistory <= 0 {
		maxHistory = 50
	}
	return &Memory{
		conversations: make(map[string]*Conversation),
		maxHistory:   maxHistory,
	}
}

func NewMemoryWithStore(maxHistory int, store ConversationStore) *Memory {
	mem := NewMemory(maxHistory)
	mem.store = store
	return mem
}

func (m *Memory) CreateConversation(id string) *Conversation {
	m.mu.Lock()
	defer m.mu.Unlock()

	conv := &Conversation{
		ID:        id,
		Messages:  make([]Message, 0),
		Context:   make(map[string]any),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.conversations[id] = conv

	if m.store != nil {
		_ = m.store.Save(conv)
	}

	return conv
}

func (m *Memory) GetConversation(id string) *Conversation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conv, exists := m.conversations[id]
	if !exists && m.store != nil {
		stored, err := m.store.Get(id)
		if err == nil && stored != nil {
			m.conversations[id] = stored
			return stored
		}
	}
	return conv
}

func (m *Memory) AddMessage(conversationID string, msg Message) {
	m.mu.Lock()
	defer m.mu.Unlock()

	conv, exists := m.conversations[conversationID]
	if !exists {
		return
	}

	conv.Messages = append(conv.Messages, msg)
	conv.UpdatedAt = time.Now()

	if len(conv.Messages) > m.maxHistory {
		conv.Messages = conv.Messages[len(conv.Messages)-m.maxHistory:]
	}

	if m.store != nil {
		_ = m.store.Save(conv)
	}
}

func (m *Memory) GetHistory(conversationID string, limit int) []Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conv, exists := m.conversations[conversationID]
	if !exists {
		return nil
	}

	if limit <= 0 || limit > len(conv.Messages) {
		limit = len(conv.Messages)
	}

	start := len(conv.Messages) - limit
	if start < 0 {
		start = 0
	}

	result := make([]Message, len(conv.Messages[start:]))
	copy(result, conv.Messages[start:])
	return result
}

func (m *Memory) SetWorkflow(conversationID string, wf *workflow.Workflow) {
	m.mu.Lock()
	defer m.mu.Unlock()

	conv, exists := m.conversations[conversationID]
	if !exists {
		return
	}

	conv.Workflow = wf
	conv.UpdatedAt = time.Now()
}

func (m *Memory) SetContext(conversationID string, key string, value any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	conv, exists := m.conversations[conversationID]
	if !exists {
		return
	}

	if conv.Context == nil {
		conv.Context = make(map[string]any)
	}
	conv.Context[key] = value
	conv.UpdatedAt = time.Now()
}

func (m *Memory) DeleteConversation(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.conversations, id)
	if m.store != nil {
		_ = m.store.Delete(id)
	}
}

func (m *Memory) ListConversations(limit, offset int) []*Conversation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.store != nil {
		stored, err := m.store.List(limit, offset)
		if err == nil {
			return stored
		}
	}

	result := make([]*Conversation, 0, len(m.conversations))
	for _, conv := range m.conversations {
		result = append(result, conv)
	}

	if len(result) > limit {
		result = result[:limit]
	}
	return result
}