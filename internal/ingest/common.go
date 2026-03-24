package ingest

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/phuc/coding-agent-reflection/internal/model"
)

type Inserter interface {
	Insert(ctx context.Context, i model.Interaction) error
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
