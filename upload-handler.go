package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Incoming %s request from %s\n", r.Method, r.RemoteAddr)

	ctype := r.Header.Get("Content-Type")
	if strings.HasPrefix(ctype, "multipart/") {
		// --- existing multipart code here ---
	} else {
		// treat body as raw file upload
		filename := fmt.Sprintf("upload-%d", time.Now().UnixNano())
		if strings.Contains(ctype, "jpeg") {
			filename += ".jpg"
		} else if strings.Contains(ctype, "png") {
			filename += ".png"
		}

		if err := os.MkdirAll("uploads", 0755); err != nil {
			http.Error(w, "could not create uploads dir", http.StatusInternalServerError)
			return
		}
		dstPath := filepath.Join("uploads", filename)
		dst, _ := os.Create(dstPath)
		defer dst.Close()
		written, _ := io.Copy(dst, r.Body)

		log.Printf("Saved raw upload to %s (%d bytes)\n", dstPath, written)
		fmt.Fprintf(w, "uploaded as %s\n", dstPath)
	}
}


func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", uploadHandler)

	addr := "0.0.0.0:1227"
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
