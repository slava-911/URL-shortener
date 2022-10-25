package metric

import "net/http"

const (
	URL = "/heartbeat"
)

type Handler struct {
}

// A HandlerFunc is a type that implement of handling an HTTP request.
type HandlerFunc interface {
	HandlerFunc(method, path string, handler http.HandlerFunc)
}

// Register adds the routes for the metric handler to the passed router.
func (h *Handler) Register(router HandlerFunc) {
	router.HandlerFunc(http.MethodGet, URL, h.Heartbeat)
}

func (h *Handler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(204)
}
