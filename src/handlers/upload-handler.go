package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"latex-clipboard/src/copy"
	"latex-clipboard/src/integrations"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	reqStart := time.Now()
	log.Printf("Incoming %s request from %s (Content-Type=%s)", r.Method, r.RemoteAddr, r.Header.Get("Content-Type"))

	ctype := r.Header.Get("Content-Type")
	if strings.HasPrefix(ctype, "multipart/") {
		http.Error(w, "multipart not yet implemented in this handler", http.StatusNotImplemented)
		return
	}

	// ---------- Save raw body ----------
	stageStart := time.Now()
	filename := fmt.Sprintf("upload-%d", time.Now().UnixNano())
	switch {
	case strings.Contains(ctype, "jpeg"):
		filename += ".jpg"
	case strings.Contains(ctype, "png"):
		filename += ".png"
	default:
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
	saveDur := time.Since(stageStart)
	log.Printf("Saved raw upload to %s (%d bytes) in %v", dstPath, written, saveDur)

	// ---------- Upload to Cloudinary ----------
	stageStart = time.Now()
	secureURL, err := integrations.UploadToCloudinary(dstPath)
	if err != nil {
		log.Printf("Cloudinary error: %v", err)
		http.Error(w, "cloudinary upload failed", http.StatusInternalServerError)
		return
	}
	cloudDur := time.Since(stageStart)
	log.Printf("Cloudinary upload done in %v → %s", cloudDur, secureURL)

	// ---------- Call LLM ----------
	stageStart = time.Now()
	latex, err := integrations.GenerateLatexFromImage(secureURL)
	if err != nil {
		log.Printf("OpenAI error: %v", err)
		http.Error(w, "openai processing failed", http.StatusInternalServerError)
		return
	}
	llmDur := time.Since(stageStart)
	totalDur := time.Since(reqStart)

	// ---------- Clipboard + Notification (best-effort) ----------
	if err := copy.CopyToClipboard(latex); err != nil {
		log.Printf("Clipboard error: %v", err)
	} else {
		log.Println("Copied result to clipboard")
	}
	copy.NotifyUser(fmt.Sprintf("LaTeX copied to clipboard (total %v)", totalDur))

	// ---------- Respond JSON (single write) ----------
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-Duration-ms", fmt.Sprintf("%d", totalDur.Milliseconds()))
	resp := map[string]any{
		"cloudinary_url": secureURL,
		"latex":          latex,
		"timings_ms": map[string]int64{
			"save":       saveDur.Milliseconds(),
			"cloudinary": cloudDur.Milliseconds(),
			"llm":        llmDur.Milliseconds(),
			"total":      totalDur.Milliseconds(),
		},
	}
	_ = json.NewEncoder(w).Encode(resp)

	log.Printf("Upload success → total=%v (save=%v, cloud=%v, llm=%v)", totalDur, saveDur, cloudDur, llmDur)
}
