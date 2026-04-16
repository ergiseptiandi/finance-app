package media

import (
	"errors"
	"fmt"
	"mime/multipart"
	"path"
	"strconv"
	"strings"
)

var (
	ErrNotFound     = errors.New("file not found")
	ErrInvalidInput = errors.New("invalid file input")
)

type FileStorage interface {
	SaveMultipartFile(relativeDir string, fileHeader *multipart.FileHeader) (string, error)
	Delete(publicPath string) error
}

type Service struct {
	storage FileStorage
}

func NewService(storage FileStorage) *Service {
	return &Service{storage: storage}
}

func (s *Service) Upload(userID int64, dir string, fileHeader *multipart.FileHeader, baseURL string) (UploadResult, error) {
	if fileHeader == nil {
		return UploadResult{}, ErrInvalidInput
	}

	relativeDir, err := resolveDirectory(userID, dir)
	if err != nil {
		return UploadResult{}, err
	}

	publicPath, err := s.storage.SaveMultipartFile(relativeDir, fileHeader)
	if err != nil {
		return UploadResult{}, err
	}

	return UploadResult{
		Path:     publicPath,
		URL:      buildFileURL(baseURL, publicPath),
		Dir:      relativeDir,
		FileName: fileHeader.Filename,
	}, nil
}

func (s *Service) GetURL(publicPath string, baseURL string) (UploadResult, error) {
	publicPath = normalizePublicPath(publicPath)
	if publicPath == "" {
		return UploadResult{}, ErrInvalidInput
	}

	return UploadResult{
		Path: publicPath,
		URL:  buildFileURL(baseURL, publicPath),
	}, nil
}

func (s *Service) Delete(publicPath string) error {
	publicPath = normalizePublicPath(publicPath)
	if publicPath == "" {
		return ErrInvalidInput
	}

	return s.storage.Delete(publicPath)
}

func resolveDirectory(userID int64, dir string) (string, error) {
	base := fmt.Sprintf("media/%d", userID)
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return base, nil
	}

	dir = strings.ReplaceAll(dir, "\\", "/")
	dir = path.Clean(strings.TrimPrefix(dir, "/"))
	if dir == "." || strings.Contains(dir, "..") {
		return "", ErrInvalidInput
	}

	return path.Join(base, dir), nil
}

func normalizePublicPath(publicPath string) string {
	publicPath = strings.TrimSpace(publicPath)
	if publicPath == "" {
		return ""
	}

	publicPath = strings.ReplaceAll(publicPath, "\\", "/")
	publicPath = path.Clean("/" + strings.TrimPrefix(publicPath, "/"))
	if publicPath == "/" || strings.Contains(publicPath, "..") {
		return ""
	}

	return publicPath
}

func buildFileURL(baseURL, publicPath string) string {
	publicPath = normalizePublicPath(publicPath)
	if publicPath == "" {
		return ""
	}

	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return publicPath
	}

	return baseURL + publicPath
}

func BaseURLFromRequest(scheme, host string) string {
	scheme = strings.TrimSpace(scheme)
	host = strings.TrimSpace(host)
	if scheme == "" {
		scheme = "http"
	}
	if host == "" {
		return ""
	}
	return scheme + "://" + host
}

func SchemeFromHeaderOrTLS(https bool, forwardedProto string) string {
	if forwardedProto != "" {
		parts := strings.Split(forwardedProto, ",")
		if len(parts) > 0 {
			scheme := strings.TrimSpace(parts[0])
			if scheme != "" {
				return scheme
			}
		}
	}
	if https {
		return "https"
	}
	return "http"
}

func ParseInt64(value string) (int64, error) {
	value = strings.TrimSpace(value)
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrInvalidInput
	}
	return id, nil
}
