package handlers

import (
	"net/http"
	"path/filepath"
)

type HomeHandler struct {
	TemplatesDir string
}

func NewHomeHandler(templatesDir string) *HomeHandler {
	return &HomeHandler{TemplatesDir: templatesDir}
}

func (h *HomeHandler) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, filepath.Join(h.TemplatesDir, "index.html"))
}
