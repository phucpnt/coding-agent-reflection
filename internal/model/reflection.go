package model

import (
	"time"

	"github.com/google/uuid"
)

type Reflection struct {
	ID            uuid.UUID
	Date          time.Time
	Summary       string
	ShouldDo      string
	ShouldNotDo   string
	ConfigChanges string
	CreatedAt     time.Time
}
