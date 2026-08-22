package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

type localStorage struct {
	rootDir string
}

func newLocal(rootDir string) *localStorage {
	return &localStorage{rootDir: rootDir}
}

func (l *localStorage) Save(_ context.Context, key string, r io.Reader, _ string) error {
	full := filepath.Join(l.rootDir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	f, err := os.Create(full)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}
