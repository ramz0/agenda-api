package models

import (
	"time"

	"github.com/google/uuid"
)

type AssignmentStatus string

const (
	AssignmentStatusPending  AssignmentStatus = "pending"
	AssignmentStatusApproved AssignmentStatus = "approved"
	AssignmentStatusRejected AssignmentStatus = "rejected"
)

type EventAssignment struct {
	ID          uuid.UUID        `db:"id" json:"id"`
	EventID     uuid.UUID        `db:"event_id" json:"eventId"`
	UserID      uuid.UUID        `db:"user_id" json:"userId"`
	Status      AssignmentStatus `db:"status" json:"status"`
	AssignedAt  time.Time        `db:"assigned_at" json:"assignedAt"`
	RespondedAt *time.Time       `db:"responded_at" json:"respondedAt,omitempty"`
}

type EventAssignmentWithDetails struct {
	EventAssignment
	UserName   string `db:"user_name" json:"userName"`
	UserEmail  string `db:"user_email" json:"userEmail"`
	EventTitle string `db:"event_title" json:"eventTitle"`
	EventDate  string `db:"event_date" json:"eventDate"`
}

type RespondAssignmentInput struct {
	Status AssignmentStatus `json:"status" binding:"required,oneof=approved rejected"`
}
