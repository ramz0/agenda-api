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

type EventHandler struct {
	eventRepo      *repository.EventRepository
	teamRepo       *repository.TeamRepository
	assignmentRepo *repository.AssignmentRepository
}

func NewEventHandler(eventRepo *repository.EventRepository, teamRepo *repository.TeamRepository, assignmentRepo *repository.AssignmentRepository) *EventHandler {
	return &EventHandler{eventRepo: eventRepo, teamRepo: teamRepo, assignmentRepo: assignmentRepo}
}

func (h *EventHandler) Create(c *gin.Context) {
	var input models.CreateEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	status := input.Status
	if status == "" {
		status = models.EventStatusDraft
	}

	eventType := input.Type
	if eventType == "" {
		eventType = models.EventTypePersonal
	}

	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)

	// Only admins can create team events
	if eventType == models.EventTypeTeam && userRole != models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can create team events"})
		return
	}

	// Team events require a team
	if eventType == models.EventTypeTeam && input.TeamID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Team events require a teamId"})
		return
	}

	// Verify team exists if teamId provided
	if input.TeamID != nil {
		_, err := h.teamRepo.GetByID(*input.TeamID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch team"})
			return
		}
	}

	event := &models.Event{
		ID:          uuid.New(),
		Title:       input.Title,
		Description: input.Description,
		Date:        date,
		StartTime:   input.StartTime,
		EndTime:     input.EndTime,
		Location:    input.Location,
		Capacity:    input.Capacity,
		Status:      status,
		Type:        eventType,
		TeamID:      input.TeamID,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.eventRepo.Create(event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	// If team event, create assignments for all team members
	if eventType == models.EventTypeTeam && input.TeamID != nil {
		members, err := h.teamRepo.GetMembers(*input.TeamID)
		if err == nil && len(members) > 0 {
			var assignments []models.EventAssignment
			for _, member := range members {
				assignments = append(assignments, models.EventAssignment{
					ID:         uuid.New(),
					EventID:    event.ID,
					UserID:     member.UserID,
					Status:     models.AssignmentStatusPending,
					AssignedAt: time.Now(),
				})
			}
			h.assignmentRepo.CreateBatch(assignments)
		}
	}

	// Set participants for personal events
	if len(input.ParticipantIds) > 0 {
		h.eventRepo.SetParticipants(event.ID, input.ParticipantIds)
	}

	// Return event with participants
	eventWithParticipants, err := h.eventRepo.GetByIDWithParticipants(event.ID)
	if err != nil {
		c.JSON(http.StatusCreated, event)
		return
	}
	c.JSON(http.StatusCreated, eventWithParticipants)
}

func (h *EventHandler) GetAll(c *gin.Context) {
	var status *models.EventStatus
	if s := c.Query("status"); s != "" {
		st := models.EventStatus(s)
		status = &st
	}

	events, err := h.eventRepo.GetAll(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	if events == nil {
		events = []models.EventWithAttendeeCount{}
	}

	c.JSON(http.StatusOK, events)
}

func (h *EventHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	event, err := h.eventRepo.GetByIDWithParticipants(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch event"})
		return
	}

	c.JSON(http.StatusOK, event)
}

func (h *EventHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	event, err := h.eventRepo.GetByID(id)
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

	// Admins can edit any event, users can only edit their personal events
	if userRole != models.RoleAdmin && event.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own events"})
		return
	}

	var input models.UpdateEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Title != nil {
		event.Title = *input.Title
	}
	if input.Description != nil {
		event.Description = *input.Description
	}
	if input.Date != nil {
		date, err := time.Parse("2006-01-02", *input.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}
		event.Date = date
	}
	if input.StartTime != nil {
		event.StartTime = *input.StartTime
	}
	if input.EndTime != nil {
		event.EndTime = *input.EndTime
	}
	if input.Location != nil {
		event.Location = *input.Location
	}
	if input.Capacity != nil {
		event.Capacity = input.Capacity
	}
	if input.Status != nil {
		event.Status = *input.Status
	}

	if err := h.eventRepo.Update(event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	// Update participants if provided
	if input.ParticipantIds != nil {
		h.eventRepo.SetParticipants(event.ID, input.ParticipantIds)
	}

	// Return event with participants
	eventWithParticipants, err := h.eventRepo.GetByIDWithParticipants(event.ID)
	if err != nil {
		c.JSON(http.StatusOK, event)
		return
	}
	c.JSON(http.StatusOK, eventWithParticipants)
}

func (h *EventHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	event, err := h.eventRepo.GetByID(id)
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

	// Admins can delete any event, users can only delete their own events
	if userRole != models.RoleAdmin && event.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own events"})
		return
	}

	if err := h.eventRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}

func (h *EventHandler) GetCalendar(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")

	if startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start and end query parameters are required"})
		return
	}

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format. Use YYYY-MM-DD"})
		return
	}

	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format. Use YYYY-MM-DD"})
		return
	}

	events, err := h.eventRepo.GetByDateRange(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	if events == nil {
		events = []models.EventWithAttendeeCount{}
	}

	c.JSON(http.StatusOK, events)
}

func (h *EventHandler) GetMyCalendar(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")

	if startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start and end query parameters are required"})
		return
	}

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format. Use YYYY-MM-DD"})
		return
	}

	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format. Use YYYY-MM-DD"})
		return
	}

	userID := middleware.GetUserID(c)

	events, err := h.eventRepo.GetCalendarByUserID(userID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	if events == nil {
		events = []models.EventWithAssignment{}
	}

	c.JSON(http.StatusOK, events)
}

func (h *EventHandler) GetMyEvents(c *gin.Context) {
	userID := middleware.GetUserID(c)
	eventType := c.Query("type")

	if eventType == "personal" {
		events, err := h.eventRepo.GetPersonalByUserID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
			return
		}
		if events == nil {
			events = []models.EventWithParticipantCount{}
		}
		c.JSON(http.StatusOK, events)
		return
	}

	if eventType == "team" {
		events, err := h.eventRepo.GetTeamEventsByUserID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
			return
		}
		if events == nil {
			events = []models.EventWithAssignmentAndCount{}
		}
		c.JSON(http.StatusOK, events)
		return
	}

	// Return all events for the user
	personalEvents, _ := h.eventRepo.GetPersonalByUserID(userID)
	teamEvents, _ := h.eventRepo.GetTeamEventsByUserID(userID)

	c.JSON(http.StatusOK, gin.H{
		"personal": personalEvents,
		"team":     teamEvents,
	})
}
