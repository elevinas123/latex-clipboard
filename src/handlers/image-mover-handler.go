package handlers

import (
	"fmt"
	"latex-clipboard/src/copy"
	"latex-clipboard/src/config"
	"latex-clipboard/src/utils"
	"net/http"
)

func ImageMoverHandler(w http.ResponseWriter, r *http.Request) {
	path, err := utils.SaveRequestBodyAsUpload(w, r, config.UploadDir, "")
	if err != nil {
		return
	}
	fmt.Printf("Saved image to %s\n", path)
	copy.NotifyUser("Image uploaded")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(path))
}
