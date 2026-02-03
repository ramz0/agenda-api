package models

import (
	"time"

	"github.com/google/uuid"
)

type EventStatus string

const (
	EventStatusDraft     EventStatus = "draft"
	EventStatusPublished EventStatus = "published"
	EventStatusCancelled EventStatus = "cancelled"
)

type EventType string

const (
	EventTypePersonal EventType = "personal"
	EventTypeTeam     EventType = "team"
)

type Event struct {
	ID          uuid.UUID   `db:"id" json:"id"`
	Title       string      `db:"title" json:"title"`
	Description string      `db:"description" json:"description"`
	Date        time.Time   `db:"date" json:"date"`
	StartTime   string      `db:"start_time" json:"startTime"`
	EndTime     string      `db:"end_time" json:"endTime"`
	Location    string      `db:"location" json:"location"`
	Capacity    *int        `db:"capacity" json:"capacity,omitempty"`
	Status      EventStatus `db:"status" json:"status"`
	Type        EventType   `db:"type" json:"type"`
	TeamID      *uuid.UUID  `db:"team_id" json:"teamId,omitempty"`
	CreatedBy   uuid.UUID   `db:"created_by" json:"createdBy"`
	CreatedAt   time.Time   `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time   `db:"updated_at" json:"updatedAt"`
}

type CreateEventInput struct {
	Title          string      `json:"title" binding:"required"`
	Description    string      `json:"description"`
	Date           string      `json:"date" binding:"required"`
	StartTime      string      `json:"startTime" binding:"required"`
	EndTime        string      `json:"endTime" binding:"required"`
	Location       string      `json:"location"`
	Capacity       *int        `json:"capacity"`
	Status         EventStatus `json:"status"`
	Type           EventType   `json:"type"`
	TeamID         *uuid.UUID  `json:"teamId"`
	ParticipantIds []uuid.UUID `json:"participantIds"`
}

type UpdateEventInput struct {
	Title          *string      `json:"title"`
	Description    *string      `json:"description"`
	Date           *string      `json:"date"`
	StartTime      *string      `json:"startTime"`
	EndTime        *string      `json:"endTime"`
	Location       *string      `json:"location"`
	Capacity       *int         `json:"capacity"`
	Status         *EventStatus `json:"status"`
	Type           *EventType   `json:"type"`
	TeamID         *uuid.UUID   `json:"teamId"`
	ParticipantIds []uuid.UUID  `json:"participantIds"`
}

type EventWithAttendeeCount struct {
	Event
	AttendeeCount int     `db:"attendee_count" json:"attendeeCount"`
	TeamName      *string `db:"team_name" json:"teamName,omitempty"`
}

type EventWithAssignment struct {
	Event
	AssignmentStatus *AssignmentStatus `db:"assignment_status" json:"assignmentStatus,omitempty"`
	TeamName         *string           `db:"team_name" json:"teamName,omitempty"`
}

type EventParticipant struct {
	UserID    uuid.UUID `db:"user_id" json:"userId"`
	UserName  string    `db:"user_name" json:"userName,omitempty"`
	UserEmail string    `db:"user_email" json:"userEmail,omitempty"`
}

type EventWithParticipants struct {
	Event
	AttendeeCount int                `db:"attendee_count" json:"attendeeCount"`
	TeamName      *string            `db:"team_name" json:"teamName,omitempty"`
	Participants  []EventParticipant `json:"participants,omitempty"`
}
