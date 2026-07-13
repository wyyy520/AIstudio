package common

import (
	"os"
	"path/filepath"
	"strings"
)

type MonorepoPath struct {
	Root string
}

func NewMonorepoPath(root string) *MonorepoPath {
	abs, err := filepath.Abs(root)
	if err != nil {
		abs = root
	}
	return &MonorepoPath{Root: abs}
}

func DetectMonorepoRoot() (*MonorepoPath, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "packages")); err == nil {
			return NewMonorepoPath(dir), nil
		}
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return NewMonorepoPath(dir), nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return NewMonorepoPath(wd), nil
		}
		dir = parent
	}
}

func (mp *MonorepoPath) PackageDir(name string) string {
	return filepath.Join(mp.Root, "packages", name)
}

func (mp *MonorepoPath) AppDir(name string) string {
	return filepath.Join(mp.Root, "apps", name)
}

func (mp *MonorepoPath) BackendDir() string {
	return filepath.Join(mp.Root, "Backend")
}

func (mp *MonorepoPath) ConfigDir() string {
	return filepath.Join(mp.Root, "Config")
}

func (mp *MonorepoPath) Rel(target string) (string, error) {
	return filepath.Rel(mp.Root, target)
}

func NormalizePath(p string) string {
	return strings.ReplaceAll(p, "\\", "/")
}
