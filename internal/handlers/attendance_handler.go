package handlers

import (
	"agenda-api/internal/middleware"
	"agenda-api/internal/models"
	"agenda-api/internal/repository"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AttendanceHandler struct {
	attendanceRepo *repository.AttendanceRepository
	eventRepo      *repository.EventRepository
}

func NewAttendanceHandler(attendanceRepo *repository.AttendanceRepository, eventRepo *repository.EventRepository) *AttendanceHandler {
	return &AttendanceHandler{
		attendanceRepo: attendanceRepo,
		eventRepo:      eventRepo,
	}
}

func (h *AttendanceHandler) Register(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	userID := middleware.GetUserID(c)

	event, err := h.eventRepo.GetByID(eventID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch event"})
		return
	}

	if event.Status != models.EventStatusPublished {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot register for unpublished event"})
		return
	}

	existing, err := h.attendanceRepo.GetByEventAndUser(eventID, userID)
	if err == nil && existing.Status == models.AttendanceStatusRegistered {
		c.JSON(http.StatusConflict, gin.H{"error": "Already registered for this event"})
		return
	}

	count, err := h.attendanceRepo.CountByEventID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check capacity"})
		return
	}

	if count >= event.Capacity {
		c.JSON(http.StatusConflict, gin.H{"error": "Event is at full capacity"})
		return
	}

	if existing != nil && existing.Status == models.AttendanceStatusCancelled {
		if err := h.attendanceRepo.UpdateStatus(existing.ID, models.AttendanceStatusRegistered); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update registration"})
			return
		}
		existing.Status = models.AttendanceStatusRegistered
		c.JSON(http.StatusOK, existing)
		return
	}

	attendance := &models.Attendance{
		ID:        uuid.New(),
		EventID:   eventID,
		UserID:    userID,
		Status:    models.AttendanceStatusRegistered,
		CreatedAt: time.Now(),
	}

	if err := h.attendanceRepo.Create(attendance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register for event"})
		return
	}

	c.JSON(http.StatusCreated, attendance)
}

func (h *AttendanceHandler) Cancel(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	userID := middleware.GetUserID(c)

	attendance, err := h.attendanceRepo.GetByEventAndUser(eventID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Registration not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch registration"})
		return
	}

	if err := h.attendanceRepo.UpdateStatus(attendance.ID, models.AttendanceStatusCancelled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel registration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Registration cancelled successfully"})
}

func (h *AttendanceHandler) GetAttendees(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	event, err := h.eventRepo.GetByID(eventID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch event"})
		return
	}

	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)

	if userRole != models.RoleAdmin && (event.SpeakerID == nil || *event.SpeakerID != userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only view attendees for your events"})
		return
	}

	attendees, err := h.attendanceRepo.GetByEventID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch attendees"})
		return
	}

	c.JSON(http.StatusOK, attendees)
}

func (h *AttendanceHandler) GetMyRegistrations(c *gin.Context) {
	userID := middleware.GetUserID(c)

	registrations, err := h.attendanceRepo.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch registrations"})
		return
	}

	c.JSON(http.StatusOK, registrations)
}
