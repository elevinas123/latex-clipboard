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

	finalPath, err := utils.ConvertHEICtoJPG(path)
	if err != nil {
		http.Error(w, "failed to convert image", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Saved image to %s\n", finalPath)
	copy.NotifyUser("Image uploaded")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(finalPath))
}
