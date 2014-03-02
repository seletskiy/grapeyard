package httpapi

import(
	"os"
	"fmt"
	"strconv"
	"net/http"
)

type Handler struct {
	version int
	pathToRelease string
	host string
}

func (h *Handler) UploadImage(version int, path string) {
	h.version = version
	h.pathToRelease = path
}

func (h *Handler) GetImageURI() string {
		return fmt.Sprintf("http://%s/api/v1/dl/rape?v=%d",
		h.host, h.version)
}

func Start(port int) *Handler {
	hostname, _ := os.Hostname()
	handler := &Handler{
		host: fmt.Sprintf("%s:%d", hostname, port),
	}

	http.HandleFunc("/api/v1/dl/rape",
		func (w http.ResponseWriter, r *http.Request) {
			getBinary(w, r, handler)
		})

	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	return handler
}

func getBinary(w http.ResponseWriter, r *http.Request, handler *Handler) {
	q := r.URL.Query()
	version, _ := strconv.Atoi(q["v"][0])
	if version == handler.version {
		http.ServeFile(w, r, handler.pathToRelease)
	} else {
		http.NotFound(w, r)
	}
}
