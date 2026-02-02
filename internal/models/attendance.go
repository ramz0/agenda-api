package models

import (
	"time"

	"github.com/google/uuid"
)

type AttendanceStatus string

const (
	AttendanceStatusRegistered AttendanceStatus = "registered"
	AttendanceStatusCancelled  AttendanceStatus = "cancelled"
	AttendanceStatusAttended   AttendanceStatus = "attended"
)

type Attendance struct {
	ID        uuid.UUID        `db:"id" json:"id"`
	EventID   uuid.UUID        `db:"event_id" json:"eventId"`
	UserID    uuid.UUID        `db:"user_id" json:"userId"`
	Status    AttendanceStatus `db:"status" json:"status"`
	CreatedAt time.Time        `db:"created_at" json:"createdAt"`
}

type AttendanceWithUser struct {
	Attendance
	UserName  string `db:"user_name" json:"userName"`
	UserEmail string `db:"user_email" json:"userEmail"`
}
