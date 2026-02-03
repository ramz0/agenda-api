package models

import (
	"time"

	"github.com/google/uuid"
)

type Team struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	CreatedBy   uuid.UUID `db:"created_by" json:"createdBy"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at" json:"updatedAt"`
}

type TeamMember struct {
	ID        uuid.UUID `db:"id" json:"id"`
	TeamID    uuid.UUID `db:"team_id" json:"teamId"`
	UserID    uuid.UUID `db:"user_id" json:"userId"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type CreateTeamInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateTeamInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type AddTeamMemberInput struct {
	UserID uuid.UUID `json:"userId" binding:"required"`
}

type TeamWithMembers struct {
	Team
	Members []TeamMemberWithUser `json:"members"`
}

type TeamMemberWithUser struct {
	TeamMember
	UserName  string `db:"user_name" json:"userName"`
	UserEmail string `db:"user_email" json:"userEmail"`
}
