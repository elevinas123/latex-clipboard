package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type CloudinaryResp struct {
	SecureURL string `json:"secure_url"`
}

var (
	CloudName    string
	UploadPreset string
)

func UploadToCloudinary(filePath string) (string, error) {
	log.Printf("Uploading %s to Cloudinary (preset=%s, cloud=%s)", filePath, UploadPreset, CloudName)

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file error: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("create form file error: %w", err)
	}
	if _, err = io.Copy(part, file); err != nil {
		return "", fmt.Errorf("copy file error: %w", err)
	}

	if err = writer.WriteField("upload_preset", UploadPreset); err != nil {
		return "", fmt.Errorf("write field error: %w", err)
	}
	writer.Close()

	url := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", CloudName)
	req, _ := http.NewRequest("POST", url, &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("Cloudinary response status: %s", resp.Status)

	if resp.StatusCode != http.StatusOK {
		log.Printf("Cloudinary response body: %s", string(body))
		return "", fmt.Errorf("cloudinary upload failed: %s", resp.Status)
	}

	var cr CloudinaryResp
	if err := json.Unmarshal(body, &cr); err != nil {
		return "", fmt.Errorf("parse response error: %w", err)
	}
	return cr.SecureURL, nil
}
