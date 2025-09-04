package main

import (
	"latex-clipboard/src/endpoints"
	"latex-clipboard/src/integrations"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var (
	cloudName    string
	uploadPreset string
)

func main() {
	// Load .env file if present (fallback to system env)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, falling back to system environment")
	}

	// Load config from env
	cloudURL := os.Getenv("CLOUDINARY_URL")
	uploadPreset = os.Getenv("CLOUDINARY_UPLOAD_PRESET")
	if cloudURL == "" || uploadPreset == "" {
		log.Fatal("CLOUDINARY_URL and CLOUDINARY_UPLOAD_PRESET must be set")
	}

	// CLOUDINARY_URL looks like: cloudinary://<api_key>:<api_secret>@<cloud_name>
	parts := strings.Split(cloudURL, "@")
	if len(parts) != 2 {
		log.Fatal("Invalid CLOUDINARY_URL format")
	}
	cloudName = parts[1]

	log.Printf("Cloudinary configured for cloud: %s with preset: %s", cloudName, uploadPreset)
	integrations.CloudName = cloudName
	integrations.UploadPreset = uploadPreset
	// Setup routes
	mux := http.NewServeMux()
	endpoints.RegisterEndpoints(mux)

	addr := "0.0.0.0:1227"
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
