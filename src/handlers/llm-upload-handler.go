package handlers

import (
	"encoding/json"
	"fmt"
	"latex-clipboard/src/copy"
	"latex-clipboard/src/integrations"
	"latex-clipboard/src/utils"
	"log"
	"net/http"
	"time"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	reqStart := time.Now()

	ctype, err := utils.ExtractAndNormalizeContentType(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	saveStart := time.Now()
	ext := utils.InferImageExtension(ctype)
	if err := utils.EnsureDir("uploads"); err != nil {
		http.Error(w, "could not create upload dir", http.StatusInternalServerError)
		return
	}
	dstPath := utils.BuildFilePath("uploads", "", ext)
	dst, err := utils.CreateFileAt(dstPath)
	if err != nil {
		http.Error(w, "could not create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	written, err := utils.CopyToFile(dst, r.Body)
	if err != nil {
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}
	saveDur := time.Since(saveStart)
	log.Printf("Saved raw upload to %s (%d bytes) in %v", dstPath, written, saveDur)

	cloudStart := time.Now()
	secureURL, err := integrations.UploadToCloudinary(dstPath)
	if err != nil {
		log.Printf("Cloudinary error: %v", err)
		http.Error(w, "cloudinary upload failed", http.StatusInternalServerError)
		return
	}
	cloudDur := time.Since(cloudStart)

	llmStart := time.Now()
	latex, err := integrations.GenerateLatexFromImage(secureURL)
	if err != nil {
		log.Printf("OpenAI error: %v", err)
		http.Error(w, "openai processing failed", http.StatusInternalServerError)
		return
	}
	llmDur := time.Since(llmStart)

	if err := copy.CopyToClipboard(latex); err != nil {
		log.Printf("Clipboard error: %v", err)
	} else {
		log.Println("Copied result to clipboard")
	}
	copy.NotifyUser(fmt.Sprintf("LaTeX copied to clipboard (total %v)", time.Since(reqStart)))

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Save-ms", fmt.Sprintf("%d", saveDur.Milliseconds()))
	w.Header().Set("X-Cloudinary-ms", fmt.Sprintf("%d", cloudDur.Milliseconds()))
	w.Header().Set("X-LLM-ms", fmt.Sprintf("%d", llmDur.Milliseconds()))
	w.Header().Set("X-Total-ms", fmt.Sprintf("%d", time.Since(reqStart).Milliseconds()))

	resp := map[string]any{
		"cloudinary_url": secureURL,
		"latex":          latex,
		"timings_ms": map[string]int64{
			"save":       saveDur.Milliseconds(),
			"cloudinary": cloudDur.Milliseconds(),
			"llm":        llmDur.Milliseconds(),
			"total":      time.Since(reqStart).Milliseconds(),
		},
	}
	_ = json.NewEncoder(w).Encode(resp)
}
