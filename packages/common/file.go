package common

import (
	"io"
	"os"
	"path/filepath"
)

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func EnsureFileDir(path string) error {
	return EnsureDir(filepath.Dir(path))
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func AtomicWrite(path string, data []byte, perm os.FileMode) error {
	if err := EnsureFileDir(path); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func CopyFile(src, dst string) error {
	if err := EnsureFileDir(dst); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
