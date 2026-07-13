package common

import (
	"time"

	"github.com/google/uuid"
)

type ID string

func NewID() ID {
	return ID(uuid.New().String())
}

func (id ID) String() string {
	return string(id)
}

func (id ID) IsZero() bool {
	return id == ""
}

type Version string

type Timestamp struct {
	time.Time
}

func NewTimestamp() Timestamp {
	return Timestamp{Time: time.Now()}
}

func (ts Timestamp) MarshalJSON() ([]byte, error) {
	return ts.Time.MarshalJSON()
}

func (ts *Timestamp) UnmarshalJSON(data []byte) error {
	return ts.Time.UnmarshalJSON(data)
}

type Metadata map[string]interface{}

func (m Metadata) Get(key string) interface{} {
	if m == nil {
		return nil
	}
	return m[key]
}

func (m Metadata) Set(key string, value interface{}) {
	m[key] = value
}

type ModuleInfo struct {
	Name    string  `json:"name"`
	Version Version `json:"version"`
	Path    string  `json:"path"`
}
