package project

import "fmt"

func DefaultGitIgnore() string {
	return `# AIStudio Project
.aistudio/
*.pyc
__pycache__/
*.so
*.dll
*.dylib
.DS_Store
Thumbs.db
node_modules/
.env
venv/
.venv/
*.egg-info/
dist/
build/
*.log
logs/
cache/
`
}

func DefaultReadme(name, description string) string {
	readme := "# " + name + "\n"
	if description != "" {
		readme += "\n" + description + "\n"
	}
	readme += fmt.Sprintf(`
## Overview
AIStudio project generated on template.

## Structure
- workflow.json - Workflow definition (Single Source of Truth)
- settings.json - Project settings
- logs/ - Runtime logs
- models/ - Trained models
- cache/ - Temporary cache

## Getting Started
1. Open in AIStudio
2. Configure your workflow
3. Build and run
`)
	return readme
}

func DefaultSettings() string {
	return `{
  "version": "1.0.0",
  "editor": {
    "theme": "default",
    "fontSize": 14
  },
  "build": {
    "autoCompile": true,
    "outputDir": "dist"
  }
}
`
}
