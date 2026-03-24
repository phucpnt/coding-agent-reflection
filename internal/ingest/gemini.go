package ingest

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/phuc/coding-agent-reflection/internal/model"
)

type geminiPayload struct {
	SessionID  string          `json:"session_id"`
	Cwd        string          `json:"cwd"`
	UserPrompt string          `json:"user_prompt"`
	Output     string          `json:"agent_output"`
	ToolsUsed  json.RawMessage `json:"tools_used"`
	Context    json.RawMessage `json:"context"`
}

func GeminiHandler(s Inserter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p geminiPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		if p.SessionID == "" {
			writeError(w, http.StatusBadRequest, "missing required field: session_id")
			return
		}

		i := model.Interaction{
			ID:          uuid.New(),
			Ts:          time.Now(),
			Provider:    "gemini",
			SessionID:   p.SessionID,
			Project:     p.Cwd,
			UserPrompt:  p.UserPrompt,
			AgentOutput: p.Output,
			Context:     string(p.Context),
		}
		if len(p.ToolsUsed) > 0 {
			i.ToolsUsed = sql.NullString{String: string(p.ToolsUsed), Valid: true}
		}

		if err := s.Insert(r.Context(), i); err != nil {
			slog.Error("insert gemini interaction", "err", err)
			writeError(w, http.StatusInternalServerError, "storage error")
			return
		}

		slog.Info("ingested gemini interaction", "session", p.SessionID)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
