// Package compiler provides the Template Engine for project generation.
//
// Template Engine uses Go's text/template to process project templates.
// All templates use double-brace syntax {{.Variable}} consistent with Go Template.
//
// Design Principles:
// - Template-driven generation (no string concatenation)
// - Domain-specific templates (Python, MATLAB, STM32, ANSYS)
// - Variable substitution from Execution Plan
// - Extensible template registry
package compiler

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// ============================================================================
// Template Engine Types
// ============================================================================

// Template represents a project template.
type Template struct {
	Name        string            `json:"name"`
	Domain      string            `json:"domain"`
	Path        string            `json:"path"`
	Description string            `json:"description"`
	Version     string            `json:"version"`
	Variables   map[string]string `json:"variables,omitempty"`
}

// TemplateEngine processes project templates using Go's text/template.
type TemplateEngine struct {
	templates map[string]*template.Template
	funcMap   template.FuncMap
}

// NewTemplateEngine creates a new TemplateEngine.
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		templates: make(map[string]*template.Template),
		funcMap: template.FuncMap{
			"upper":     strings.ToUpper,
			"lower":     strings.ToLower,
			"title":     strings.Title,
			"trim":      strings.TrimSpace,
			"replace":   strings.ReplaceAll,
			"join":      strings.Join,
			"split":     strings.Split,
			"contains":  strings.Contains,
			"hasPrefix": strings.HasPrefix,
			"hasSuffix": strings.HasSuffix,
		},
	}
}

// RegisterTemplate registers a template from a directory.
func (e *TemplateEngine) RegisterTemplate(name, dir string) error {
	tmpl := template.New(name).Funcs(e.funcMap)

	// Parse all template files in the directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Only process template files
		ext := filepath.Ext(path)
		if ext == ".go.tmpl" || ext == ".tmpl" || ext == ".template" {
			relPath, _ := filepath.Rel(dir, path)
			_, err := tmpl.ParseFiles(path)
			if err != nil {
				return fmt.Errorf("parse template %s: %w", relPath, err)
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	e.templates[name] = tmpl
	return nil
}

// Render renders a template with the given variables.
func (e *TemplateEngine) Render(name string, data map[string]interface{}) (string, error) {
	tmpl, ok := e.templates[name]
	if !ok {
		return "", fmt.Errorf("template %q not found", name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template %q: %w", name, err)
	}

	return buf.String(), nil
}

// RenderFile renders a template file and writes the output.
func (e *TemplateEngine) RenderFile(name string, data map[string]interface{}, outputPath string) error {
	content, err := e.Render(name, data)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	return os.WriteFile(outputPath, []byte(content), 0644)
}

// ProcessTemplateDir processes an entire template directory, rendering all template files.
func (e *TemplateEngine) ProcessTemplateDir(templateDir string, outputDir string, data map[string]interface{}) error {
	return filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(templateDir, path)
		if err != nil {
			return err
		}

		outputPath := filepath.Join(outputDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(outputPath, 0755)
		}

		// Check if file is a template
		ext := filepath.Ext(path)
		isTemplate := ext == ".go.tmpl" || ext == ".tmpl" || ext == ".template"

		if isTemplate {
			// Read template content
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Create template and render
			tmplName := filepath.Base(path)
			tmpl, err := template.New(tmplName).Funcs(e.funcMap).Parse(string(content))
			if err != nil {
				return fmt.Errorf("parse template %s: %w", relPath, err)
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, data); err != nil {
				return fmt.Errorf("execute template %s: %w", relPath, err)
			}

			// Write output (remove template extension)
			outputPath = strings.TrimSuffix(outputPath, ext)
			if ext == ".go.tmpl" {
				outputPath = strings.TrimSuffix(outputPath, ".go")
			}
			return os.WriteFile(outputPath, buf.Bytes(), 0644)
		}

		// Copy non-template files as-is
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(outputPath, content, info.Mode())
	})
}

// ListTemplates returns all registered template names.
func (e *TemplateEngine) ListTemplates() []string {
	names := make([]string, 0, len(e.templates))
	for name := range e.templates {
		names = append(names, name)
	}
	return names
}

// HasTemplate checks if a template is registered.
func (e *TemplateEngine) HasTemplate(name string) bool {
	_, ok := e.templates[name]
	return ok
}
