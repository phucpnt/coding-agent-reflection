package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Interaction struct {
	ID           uuid.UUID
	Ts           time.Time
	Provider     string
	SessionID    string
	Project      string
	UserPrompt   string
	AgentOutput  string
	Context      string
	TokensPrompt sql.NullInt64
	TokensOutput sql.NullInt64
	ToolsUsed    sql.NullString
}
