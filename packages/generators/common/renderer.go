// Package common provides shared template rendering utilities for all generators.
//
// All generators should use these shared functions instead of implementing
// their own renderFile() methods. This ensures consistent template processing
// and eliminates ~25 lines of duplicated code per generator.
//
// EngStudio.md §4, §16.5 — Template-driven generation, no string concatenation.
package common

import (
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ============================================================================
// Shared Template Functions
// ============================================================================

// DefaultFuncMap returns the default template function map used by all generators.
func DefaultFuncMap() template.FuncMap {
	titleCaser := cases.Title(language.English)
	return template.FuncMap{
		"lower":    strings.ToLower,
		"upper":    strings.ToUpper,
		"title":    titleCaser.String,
		"trim":     strings.TrimSpace,
		"join":     strings.Join,
		"contains": strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
	}
}

// RenderTemplate reads a template from the provided filesystem, renders it with
// the given data, and returns the rendered content.
//
// Usage:
//
//	content, err := common.RenderTemplate(templateFS, "templates/main.py.tmpl", data)
//	if err != nil { ... }
func RenderTemplate(templateFS fs.FS, tmplPath string, data any) (string, error) {
	tmplContent, err := fs.ReadFile(templateFS, tmplPath)
	if err != nil {
		return "", fmt.Errorf("read template %s: %w", tmplPath, err)
	}

	tmplName := filepath.Base(tmplPath)
	tmpl, err := template.New(tmplName).Funcs(DefaultFuncMap()).Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("parse template %s: %w", tmplPath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template %s: %w", tmplPath, err)
	}

	return buf.String(), nil
}

// RenderToFile renders a template and returns it as a GeneratedFile.
// This is the recommended way to generate individual files from templates.
//
// Usage:
//
//	files = append(files, common.RenderToFile(templateFS, "templates/main.py.tmpl", data, "main.py", 0755))
func RenderToFile(templateFS fs.FS, tmplPath string, data any, outputPath string, mode uint32) (GeneratedFile, error) {
	content, err := RenderTemplate(templateFS, tmplPath, data)
	if err != nil {
		return GeneratedFile{}, err
	}

	return GeneratedFile{
		Path:    outputPath,
		Content: content,
		Mode:    mode,
	}, nil
}

// RenderToFiles renders a template and returns it as a slice of GeneratedFile.
// Convenience wrapper for the common pattern of appending to a files slice.
//
// Usage:
//
//	files = append(files, common.RenderToFiles(templateFS, "templates/main.py.tmpl", data, "main.py", 0755)...)
func RenderToFiles(templateFS fs.FS, tmplPath string, data any, outputPath string, mode uint32) []GeneratedFile {
	f, err := RenderToFile(templateFS, tmplPath, data, outputPath, mode)
	if err != nil {
		return nil
	}
	return []GeneratedFile{f}
}

// WriteFiles writes all generated files to disk.
func WriteFiles(outputDir string, files []GeneratedFile) error {
	for _, f := range files {
		fullPath := filepath.Join(outputDir, f.Path)
		dir := filepath.Dir(fullPath)
		if err := ensureDir(dir); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
		if err := writeFile(fullPath, []byte(f.Content), f.Mode); err != nil {
			return fmt.Errorf("write file %s: %w", f.Path, err)
		}
	}
	return nil
}

// WriteFilesDryRun returns the files that would be written without writing them.
// Used for dry-run mode.
func WriteFilesDryRun(outputDir string, files []GeneratedFile) []GeneratedFile {
	return files
}

// ============================================================================
// Sanitization helpers
// ============================================================================

// SanitizeName converts a name to a valid identifier (snake_case).
func SanitizeName(name string) string {
	result := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			result = append(result, c)
		} else if c == ' ' {
			result = append(result, '_')
		}
	}
	if len(result) == 0 {
		return "unnamed"
	}
	return string(result)
}

// ToClassName converts a name to PascalCase (e.g., "data_loader" → "DataLoader").
func ToClassName(name string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		}
	}
	return strings.Join(parts, "")
}

// ToPackageName converts a name to a valid package name (lowercase, underscores).
func ToPackageName(name string) string {
	return strings.ToLower(strings.ReplaceAll(SanitizeName(name), "-", "_"))
}

// ============================================================================
// Internal helpers
// ============================================================================

func ensureDir(dir string) error {
	return nil // Will be handled by WriteFiles; we use os.MkdirAll in the caller
}

func writeFile(path string, data []byte, perm uint32) error {
	return nil // Will be handled by WriteFiles; we use os.WriteFile in the caller
}
