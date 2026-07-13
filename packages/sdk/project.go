package sdk

import (
	"github.com/aistudio/packages/project"
)

type Project = project.Project
type ProjectConfig = project.ProjectConfig

func CreateProject(name, target, dir string) (*Project, error) {
	mgr := project.NewManager(dir)
	return mgr.Create(name, target, dir)
}

func OpenProject(path string) (*Project, error) {
	mgr := project.NewManager("")
	return mgr.Open(path)
}

func ImportProject(path string) (*Project, error) {
	mgr := project.NewManager("")
	return mgr.Import(path)
}