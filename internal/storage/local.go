package storage

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LocalStorage struct {
	BaseDir string
}

func NewLocalStorage(baseDir string) *LocalStorage {
	return &LocalStorage{BaseDir: baseDir}
}

func (s *LocalStorage) SaveMultipartFile(relativeDir string, fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader == nil {
		return "", errors.New("proof_image is required")
	}

	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dir := filepath.Join(s.BaseDir, filepath.FromSlash(relativeDir))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	ext := filepath.Ext(fileHeader.Filename)
	base := strings.TrimSuffix(filepath.Base(fileHeader.Filename), ext)
	safeBase := sanitizeFilename(base)
	fileName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), safeBase, ext)

	fullPath := filepath.Join(dir, fileName)
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	publicPath := "/" + filepath.ToSlash(filepath.Join(relativeDir, fileName))
	return publicPath, nil
}

func (s *LocalStorage) Delete(publicPath string) error {
	trimmed := strings.TrimPrefix(publicPath, "/")
	if trimmed == "" {
		return nil
	}

	fullPath := filepath.Join(s.BaseDir, filepath.FromSlash(trimmed))
	if err := os.Remove(fullPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func sanitizeFilename(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "file"
	}

	builder := strings.Builder{}
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '-' || r == '_':
			builder.WriteRune(r)
		default:
			builder.WriteRune('_')
		}
	}

	return strings.Trim(builder.String(), "_")
}
