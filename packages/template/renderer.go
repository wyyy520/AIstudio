package template

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// RenderedFile represents a single rendered output file.
type RenderedFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Mode    uint32 `json:"mode"`
}

// RenderOptions controls the template rendering behavior.
type RenderOptions struct {
	Variables  map[string]string // {{var}} replacements
	OutputDir  string            // where to write rendered files
	Force      bool              // overwrite existing output
	DryRun     bool              // validate without writing
	PluginDirs []string          // extra directories to search for template files
}

// Render copies a template directory and substitutes variables.
// It processes all files in the template directory:
//   - .tpl files are parsed as Go templates, executed with variable map, and written without .tpl extension
//   - All other files are copied as-is
func (e *Engine) Render(templateID string, outputDir string, vars map[string]string) ([]RenderedFile, error) {
	tmpl, ok := e.Get(templateID)
	if !ok {
		return nil, fmt.Errorf("template %q not found", templateID)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("create output directory %s: %w", outputDir, err)
	}

	var files []RenderedFile

	err := filepath.Walk(tmpl.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable
		}
		if info.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			// Create corresponding output directory
			rel, _ := filepath.Rel(tmpl.Path, path)
			if rel != "." {
				dir := filepath.Join(outputDir, rel)
				if err := os.MkdirAll(dir, 0755); err != nil {
					return err
				}
			}
			return nil
		}

		// Skip template.json metadata file
		if info.Name() == "template.json" {
			return nil
		}

		rel, _ := filepath.Rel(tmpl.Path, path)
		if strings.HasPrefix(rel, ".") {
			return nil // skip hidden files
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable
		}

		outputPath := filepath.Join(outputDir, rel)
		var renderedContent string

		if strings.HasSuffix(info.Name(), ".tpl") {
			// Process as Go template
			outputPath = strings.TrimSuffix(outputPath, ".tpl")
			renderedContent, err = e.renderTemplate(string(content), vars)
			if err != nil {
				return fmt.Errorf("render %s: %w", rel, err)
			}
		} else {
			// Copy as-is with variable substitution for non-binary files
			if isTextFile(info.Name()) {
				renderedContent = e.substituteVars(string(content), vars)
			} else {
				renderedContent = string(content)
			}
		}

		// Create parent directories for the output file
		dir := filepath.Dir(outputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}

		files = append(files, RenderedFile{
			Path:    rel,
			Content: renderedContent,
			Mode:    uint32(info.Mode().Perm()),
		})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk template directory: %w", err)
	}

	return files, nil
}

// renderTemplate parses a Go template string and executes it with the variable map.
func (e *Engine) renderTemplate(content string, vars map[string]string) (string, error) {
	funcMap := template.FuncMap{
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		"title": strings.Title,
		"default": func(s, def string) string {
			if s == "" {
				return def
			}
			return s
		},
	}

	tmpl, err := template.New("render").Funcs(funcMap).Parse(content)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

// substituteVars replaces {{var}} patterns with values from the variable map.
// This is a simpler substitution for non-template files that still contain {{var}} references.
func (e *Engine) substituteVars(content string, vars map[string]string) string {
	for k, v := range vars {
		content = strings.ReplaceAll(content, "{{"+k+"}}", v)
	}
	return content
}

// isTextFile returns true if the file extension suggests a text file.
func isTextFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".go", ".py", ".js", ".ts", ".vue", ".html", ".css", ".yaml", ".yml",
		".json", ".xml", ".md", ".txt", ".toml", ".ini", ".cfg", ".conf",
		".sh", ".bat", ".ps1", ".env", ".gitignore", ".dockerignore",
		".c", ".h", ".cpp", ".hpp", ".java", ".rs", ".matlab", ".m":
		return true
	}
	return false
}

// FlattenVariables converts a map[string]any to map[string]string for template rendering.
func FlattenVariables(vars map[string]any) map[string]string {
	result := make(map[string]string, len(vars))
	for k, v := range vars {
		switch val := v.(type) {
		case string:
			result[k] = val
		case fmt.Stringer:
			result[k] = val.String()
		default:
			result[k] = fmt.Sprintf("%v", v)
		}
	}
	return result
}

// jsonUnmarshal handles JSON unmarshaling.
func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
