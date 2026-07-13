package skill

import (
	"time"

	"github.com/aistudio/packages/workflow"
)

type Category string

const (
	CategoryObjectDetection  Category = "object_detection"
	CategoryClassification   Category = "classification"
	CategorySegmentation     Category = "segmentation"
	CategoryNLP              Category = "nlp"
	CategoryAudio            Category = "audio"
	CategorySimulation       Category = "simulation"
	CategoryDataProcessing   Category = "data_processing"
	CategoryCustom           Category = "custom"
)

type Skill struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Version         string            `json:"version"`
	Author          string            `json:"author,omitempty"`
	Tags            []string          `json:"tags,omitempty"`
	Category        Category          `json:"category"`
	MinSchemaVersion string           `json:"min_schema_version"`
	Workflow        *workflow.Workflow `json:"workflow"`
	CreatedAt       time.Time         `json:"created_at,omitempty"`
	UpdatedAt       time.Time         `json:"updated_at,omitempty"`
}

type SkillSummary struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Category    Category `json:"category"`
	Tags        []string `json:"tags,omitempty"`
	Author      string   `json:"author,omitempty"`
}