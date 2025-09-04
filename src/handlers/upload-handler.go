package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"latex-clipboard/src/integrations"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"latex-clipboard/src/copy"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Incoming %s request from %s (Content-Type=%s)", r.Method, r.RemoteAddr, r.Header.Get("Content-Type"))

	ctype := r.Header.Get("Content-Type")
	if strings.HasPrefix(ctype, "multipart/") {
		http.Error(w, "multipart not yet implemented in this handler", http.StatusNotImplemented)
		return
	}

	// raw body upload
	filename := fmt.Sprintf("upload-%d", time.Now().UnixNano())
	if strings.Contains(ctype, "jpeg") {
		filename += ".jpg"
	} else if strings.Contains(ctype, "png") {
		filename += ".png"
	} else {
		filename += ".bin"
	}

	if err := os.MkdirAll("uploads", 0755); err != nil {
		log.Printf("Error creating uploads dir: %v", err)
		http.Error(w, "could not create uploads dir", http.StatusInternalServerError)
		return
	}

	dstPath := filepath.Join("uploads", filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		http.Error(w, "could not create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, r.Body)
	if err != nil {
		log.Printf("Error writing file: %v", err)
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}

	log.Printf("Saved raw upload to %s (%d bytes)", dstPath, written)

	// Push to Cloudinary
	secureURL, err := integrations.UploadToCloudinary(dstPath)
	if err != nil {
		log.Printf("Cloudinary error: %v", err)
		http.Error(w, "cloudinary upload failed", http.StatusInternalServerError)
		return
	}

	// Call LLM
	latex, err := integrations.GenerateLatexFromImage(secureURL)
	if err != nil {
		log.Printf("OpenAI error: %v", err)
		http.Error(w, "openai processing failed", http.StatusInternalServerError)
		return
	}

	// Respond with JSON (Cloudinary URL + LaTeX)
	resp := map[string]string{
		"cloudinary_url": secureURL,
		"latex":          latex,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	// Save to clipboard
	if err := copy.CopyToClipboard(latex); err != nil {
		log.Printf("Clipboard error: %v", err)
	} else {
		log.Println("Copied result to clipboard")
	}

	// Notify user
	copy.NotifyUser("LaTeX copied to clipboard")


	log.Printf("Upload success → %s", secureURL)
	fmt.Fprintf(w, "Cloudinary URL: %s\n", secureURL)
}
