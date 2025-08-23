package api

import (
	"encoding/json"
	"net/http"

	"github.com/svaan1/go-tcc/internal/grpcserver"
)

type Handlers struct {
	sv *grpcserver.Server
}

func NewHandlers(sv *grpcserver.Server) *Handlers {
	return &Handlers{sv}
}

func (h *Handlers) GetNodes(w http.ResponseWriter, r *http.Request) {
	nodes := h.sv.GetNodes()

	data, err := json.Marshal(nodes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (h *Handlers) AssignJob(w http.ResponseWriter, r *http.Request) {
	h.sv.AssignJob("./samples/input.mp4", "./samples/output.mp4", "28", "slow", "aac", "libx264")

	w.WriteHeader(http.StatusCreated)
}
