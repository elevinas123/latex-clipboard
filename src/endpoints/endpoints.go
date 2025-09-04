package endpoints
import (
	"latex-clipboard/src/handlers"
	"net/http"
)

func RegisterEndpoints(mux *http.ServeMux) {
	mux.HandleFunc("/upload", handlers.UploadHandler)
	mux.HandleFunc("/health", handlers.HealthHandler)
}
