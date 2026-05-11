package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PersonalAccessToken struct {
	bun.BaseModel `bun:"table:personal_access_tokens,alias:pat"`

	ID        uuid.UUID `bun:"type:uuid,default:gen_random_uuid(),pk"`
	UserID    uuid.UUID `bun:"type:uuid,notnull"`
	Token     string    `bun:",unique,notnull"`
	ExpiresAt time.Time `bun:",notnull"`
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`

	User *User `bun:"rel:belongs-to,join:user_id=id"`
}
