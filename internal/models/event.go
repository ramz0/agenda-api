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

type Event struct {
	ID          uuid.UUID   `db:"id" json:"id"`
	Title       string      `db:"title" json:"title"`
	Description string      `db:"description" json:"description"`
	Date        time.Time   `db:"date" json:"date"`
	StartTime   string      `db:"start_time" json:"startTime"`
	EndTime     string      `db:"end_time" json:"endTime"`
	Location    string      `db:"location" json:"location"`
	Capacity    int         `db:"capacity" json:"capacity"`
	Status      EventStatus `db:"status" json:"status"`
	CreatedBy   uuid.UUID   `db:"created_by" json:"createdBy"`
	SpeakerID   *uuid.UUID  `db:"speaker_id" json:"speakerId,omitempty"`
	CreatedAt   time.Time   `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time   `db:"updated_at" json:"updatedAt"`
}

type CreateEventInput struct {
	Title       string      `json:"title" binding:"required"`
	Description string      `json:"description"`
	Date        string      `json:"date" binding:"required"`
	StartTime   string      `json:"startTime" binding:"required"`
	EndTime     string      `json:"endTime" binding:"required"`
	Location    string      `json:"location" binding:"required"`
	Capacity    int         `json:"capacity" binding:"required,min=1"`
	Status      EventStatus `json:"status"`
	SpeakerID   *uuid.UUID  `json:"speakerId"`
}

type UpdateEventInput struct {
	Title       *string      `json:"title"`
	Description *string      `json:"description"`
	Date        *string      `json:"date"`
	StartTime   *string      `json:"startTime"`
	EndTime     *string      `json:"endTime"`
	Location    *string      `json:"location"`
	Capacity    *int         `json:"capacity"`
	Status      *EventStatus `json:"status"`
	SpeakerID   *uuid.UUID   `json:"speakerId"`
}

type EventWithAttendeeCount struct {
	Event
	AttendeeCount int `db:"attendee_count" json:"attendeeCount"`
}
