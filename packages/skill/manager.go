package skill

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/aistudio/packages/workflow"
)

type SkillManager struct {
	mu       sync.RWMutex
	registry map[string]*Skill
}

func NewSkillManager() *SkillManager {
	return &SkillManager{
		registry: make(map[string]*Skill),
	}
}

func (m *SkillManager) Register(skill *Skill) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if skill.ID == "" {
		return fmt.Errorf("skill ID is required")
	}
	if _, exists := m.registry[skill.ID]; exists {
		return fmt.Errorf("skill already registered: %s", skill.ID)
	}
	m.registry[skill.ID] = skill
	return nil
}

func (m *SkillManager) MustRegister(skill *Skill) {
	if err := m.Register(skill); err != nil {
		panic(err)
	}
}

func (m *SkillManager) List() []SkillSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summaries := make([]SkillSummary, 0, len(m.registry))
	for _, s := range m.registry {
		summaries = append(summaries, SkillSummary{
			ID:          s.ID,
			Name:        s.Name,
			Description: s.Description,
			Version:     s.Version,
			Category:    s.Category,
			Tags:        s.Tags,
			Author:      s.Author,
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Name < summaries[j].Name
	})
	return summaries
}

func (m *SkillManager) Get(id string) *Skill {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.registry[id]
}

func (m *SkillManager) GetWorkflow(id string) *workflow.Workflow {
	m.mu.RLock()
	defer m.mu.RUnlock()

	skill, ok := m.registry[id]
	if !ok {
		return nil
	}
	return skill.Workflow
}

func (m *SkillManager) Load(path string) (*Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read skill file: %w", err)
	}

	var skill Skill
	if err := json.Unmarshal(data, &skill); err != nil {
		return nil, fmt.Errorf("failed to parse skill: %w", err)
	}

	if skill.ID == "" {
		return nil, fmt.Errorf("skill ID is required")
	}
	if skill.Name == "" {
		return nil, fmt.Errorf("skill name is required")
	}

	m.mu.Lock()
	m.registry[skill.ID] = &skill
	m.mu.Unlock()

	return &skill, nil
}

func (m *SkillManager) Search(query string) []SkillSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	query = strings.ToLower(query)
	var results []SkillSummary

	for _, s := range m.registry {
		if matches(s, query) {
			results = append(results, SkillSummary{
				ID:          s.ID,
				Name:        s.Name,
				Description: s.Description,
				Version:     s.Version,
				Category:    s.Category,
				Tags:        s.Tags,
				Author:      s.Author,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})
	return results
}

func (m *SkillManager) Categories() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	categorySet := make(map[Category]struct{})
	for _, s := range m.registry {
		categorySet[s.Category] = struct{}{}
	}

	categories := make([]string, 0, len(categorySet))
	for c := range categorySet {
		categories = append(categories, string(c))
	}
	sort.Strings(categories)
	return categories
}

func matches(s *Skill, query string) bool {
	if query == "" {
		return true
	}
	if strings.Contains(strings.ToLower(s.ID), query) {
		return true
	}
	if strings.Contains(strings.ToLower(s.Name), query) {
		return true
	}
	if strings.Contains(strings.ToLower(s.Description), query) {
		return true
	}
	if strings.Contains(strings.ToLower(string(s.Category)), query) {
		return true
	}
	for _, tag := range s.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}