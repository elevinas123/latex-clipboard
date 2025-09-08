// utils/convert.go
package utils

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

func run(cmd string, args ...string) error {
	log.Printf("Running command: %s %s", cmd, strings.Join(args, " "))
	c := exec.Command(cmd, args...)
	var stderr bytes.Buffer
	c.Stderr = &stderr

	if err := c.Run(); err != nil {
		log.Printf("%s failed: %s", cmd, strings.TrimSpace(stderr.String()))
		return fmt.Errorf("%s failed: %w", cmd, err)
	}
	log.Printf("%s succeeded", cmd)
	return nil
}

// ConvertHEICtoJPG tries multiple converters in order and logs progress.
func ConvertHEICtoJPG(inputPath string) (string, error) {
	log.Printf("Starting HEIC→JPG conversion for: %s", inputPath)

	ext := strings.ToLower(filepath.Ext(inputPath))
	if ext != ".heic" && ext != ".heif" {
		log.Printf("File %s is not HEIC/HEIF, skipping conversion", inputPath)
		return inputPath, nil
	}

	out := strings.TrimSuffix(inputPath, ext) + ".jpg"

	// Try heif-convert
	if err := run("heif-convert", inputPath, out); err == nil {
		return out, nil
	}

	// Try ffmpeg
	if err := run("ffmpeg", "-y", "-i", inputPath, "-frames:v", "1", out); err == nil {
		return out, nil
	}

	// Try ImageMagick
	if err := run("magick", inputPath+"[0]", "-strip", "-quality", "90", out); err == nil {
		return out, nil
	}

	log.Printf("All converters failed for %s", inputPath)
	return "", fmt.Errorf("all HEIC→JPG converters failed for %s", inputPath)
}
