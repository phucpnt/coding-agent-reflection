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

// OTLP JSON structures (subset we care about)
type otlpTraceRequest struct {
	ResourceSpans []resourceSpan `json:"resourceSpans"`
}

type resourceSpan struct {
	ScopeSpans []scopeSpan `json:"scopeSpans"`
}

type scopeSpan struct {
	Spans []span `json:"spans"`
}

type span struct {
	Name       string      `json:"name"`
	Attributes []attribute `json:"attributes"`
}

type attribute struct {
	Key   string         `json:"key"`
	Value attributeValue `json:"value"`
}

type attributeValue struct {
	StringValue string `json:"stringValue"`
	IntValue    string `json:"intValue"`
}

func CodexHandler(s Inserter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req otlpTraceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid OTLP JSON: "+err.Error())
			return
		}

		ingested := 0
		for _, rs := range req.ResourceSpans {
			for _, ss := range rs.ScopeSpans {
				for _, sp := range ss.Spans {
					attrs := extractAttrs(sp.Attributes)

					userPrompt, hasPrompt := attrs["user_prompt"]
					completion, hasCompletion := attrs["completion"]
					if !hasPrompt && !hasCompletion {
						slog.Debug("skipping span without LLM attributes", "name", sp.Name)
						continue
					}

					i := model.Interaction{
						ID:          uuid.New(),
						Ts:          time.Now(),
						Provider:    "codex",
						SessionID:   attrs["session_id"],
						Project:     attrs["code.filepath"],
						UserPrompt:  userPrompt,
						AgentOutput: completion,
					}

					if v, ok := attrs["code.function"]; ok {
						i.Context = `{"function":"` + v + `"}`
					}

					if err := s.Insert(r.Context(), i); err != nil {
						slog.Error("insert codex interaction", "err", err)
						writeError(w, http.StatusInternalServerError, "storage error")
						return
					}
					ingested++
				}
			}
		}

		slog.Info("ingested codex traces", "count", ingested)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"status": "ok", "ingested": ingested})
	}
}

func extractAttrs(attrs []attribute) map[string]string {
	m := make(map[string]string, len(attrs))
	for _, a := range attrs {
		if a.Value.StringValue != "" {
			m[a.Key] = a.Value.StringValue
		} else if a.Value.IntValue != "" {
			m[a.Key] = a.Value.IntValue
		}
	}
	return m
}

// ensure sql.NullString is importable (used by other files in package)
var _ = sql.NullString{}
