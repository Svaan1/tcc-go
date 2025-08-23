package api

import (
	"encoding/json"
	"net/http"

	"github.com/svaan1/go-tcc/internal/app"
	"github.com/svaan1/go-tcc/internal/grpcserver"
)

type Handlers struct {
	app *app.Service
	// temporary: we still need to push jobs over gRPC; keep a sender for now
	sender *grpcserver.Server
}

func NewHandlers(app *app.Service, sender *grpcserver.Server) *Handlers {
	return &Handlers{app: app, sender: sender}
}

func (h *Handlers) GetNodes(w http.ResponseWriter, r *http.Request) {
	nodes := h.app.ListNodes(r.Context())

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
	// App would own job decisions in a future change; for now call into transport
	h.sender.AssignJob("./samples/input.mp4", "./samples/output.mp4", "28", "slow", "aac", "libx264")

	w.WriteHeader(http.StatusCreated)
}
