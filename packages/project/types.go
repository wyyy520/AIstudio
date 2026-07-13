package project

import "time"

type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusArchived ProjectStatus = "archived"
	ProjectStatusDeleted  ProjectStatus = "deleted"
)

type ProjectConfig struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Target      string `json:"target"`
	Template    string `json:"template,omitempty"`
}

type Project struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	RootPath    string        `json:"rootPath"`
	WorkflowID  string        `json:"workflowId"`
	Target      string        `json:"target"`
	Status      ProjectStatus `json:"status"`
	FileCount   int           `json:"fileCount"`
	SizeBytes   int64         `json:"sizeBytes"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
}

type ProjectTemplate struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	WorkflowJSON string `json:"workflowJson"`
	GitIgnore    string `json:"gitIgnore"`
	Readme       string `json:"readme"`
}
