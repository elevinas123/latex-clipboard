package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func ExtractAndNormalizeContentType(r *http.Request) (string, error) {
	ctype := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	if ctype == "" {
		return "", fmt.Errorf("missing Content-Type header")
	}
	if strings.HasPrefix(ctype, "multipart/") {
		return "", fmt.Errorf("multipart not implemented")
	}
	if i := strings.Index(ctype, ";"); i >= 0 {
		ctype = strings.TrimSpace(ctype[:i])
	}
	return ctype, nil
}

func InferImageExtension(ctype string) string {
	switch {
	case strings.Contains(ctype, "jpeg"):
		return ".jpg"
	case strings.Contains(ctype, "png"):
		return ".png"
	default:
		return ".bin"
	}
}

func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

func BuildFilePath(dir, baseName, ext string) string {
	if baseName == "" {
		baseName = fmt.Sprintf("upload-%d", time.Now().UnixNano())
	}
	return filepath.Join(dir, baseName+ext)
}

func CreateFileAt(path string) (*os.File, error) {
	return os.Create(path)
}

func CopyToFile(dst *os.File, src io.Reader) (int64, error) {
	return io.Copy(dst, src)
}

func SaveRequestBodyAsUpload(w http.ResponseWriter, r *http.Request, dir, baseName string) (string, error) {
	ctype, err := ExtractAndNormalizeContentType(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return "", err
	}
	ext := InferImageExtension(ctype)
	if err := EnsureDir(dir); err != nil {
		http.Error(w, "could not create upload dir", http.StatusInternalServerError)
		return "", err
	}
	path := BuildFilePath(dir, baseName, ext)
	dst, err := CreateFileAt(path)
	if err != nil {
		http.Error(w, "could not create file", http.StatusInternalServerError)
		return "", err
	}
	defer dst.Close()
	if _, err := CopyToFile(dst, r.Body); err != nil {
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return "", err
	}
	return path, nil
}
