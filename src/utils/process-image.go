package utils

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func ExtractAndNormalizeContentTypeFromHeader(r *http.Request) string {
	ctype := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	if ctype == "" {
		return ""
	}
	if mt, _, err := mime.ParseMediaType(ctype); err == nil {
		return strings.ToLower(mt)
	}
	if i := strings.Index(ctype, ";"); i >= 0 {
		ctype = strings.TrimSpace(ctype[:i])
	}
	return ctype
}

func InferImageExtension(ctype string) string {
	ct := strings.ToLower(ctype)
	switch {
	case strings.Contains(ct, "jpeg"):
		return ".jpg"
	case strings.Contains(ct, "png"):
		return ".png"
	case strings.Contains(ct, "webp"):
		return ".webp"
	case strings.Contains(ct, "heic"),
		strings.Contains(ct, "heif"):
		return ".heic"
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
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 25<<20)

	headerCT := ExtractAndNormalizeContentTypeFromHeader(r)

	const sniffSize = 512
	buf := make([]byte, sniffSize)
	n, err := r.Body.Read(buf)
	if err != nil && err != io.EOF {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return "", fmt.Errorf("failed to read request body: %w", err)
	}
	buf = buf[:n]

	detectedCT := http.DetectContentType(buf)
	ctype := headerCT
	if ctype == "" || ctype == "application/octet-stream" {
		ctype = detectedCT
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

	reader := io.MultiReader(bytes.NewReader(buf), r.Body)
	if _, err := CopyToFile(dst, reader); err != nil {
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return "", err
	}

	return path, nil
}

