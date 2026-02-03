package handlers

import (
	"agenda-api/internal/middleware"
	"agenda-api/internal/models"
	"agenda-api/internal/repository"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AssignmentHandler struct {
	assignmentRepo *repository.AssignmentRepository
	eventRepo      *repository.EventRepository
}

func NewAssignmentHandler(assignmentRepo *repository.AssignmentRepository, eventRepo *repository.EventRepository) *AssignmentHandler {
	return &AssignmentHandler{assignmentRepo: assignmentRepo, eventRepo: eventRepo}
}

func (h *AssignmentHandler) GetMyAssignments(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var status *models.AssignmentStatus
	if s := c.Query("status"); s != "" {
		st := models.AssignmentStatus(s)
		status = &st
	}

	assignments, err := h.assignmentRepo.GetByUserID(userID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignments"})
		return
	}

	if assignments == nil {
		assignments = []models.EventAssignmentWithDetails{}
	}

	c.JSON(http.StatusOK, assignments)
}

func (h *AssignmentHandler) GetByEventID(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	assignments, err := h.assignmentRepo.GetByEventID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignments"})
		return
	}

	if assignments == nil {
		assignments = []models.EventAssignmentWithDetails{}
	}

	c.JSON(http.StatusOK, assignments)
}

func (h *AssignmentHandler) Respond(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	userID := middleware.GetUserID(c)

	// Get the assignment
	assignment, err := h.assignmentRepo.GetByEventAndUser(eventID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignment"})
		return
	}

	// Check if already responded
	if assignment.Status != models.AssignmentStatusPending {
		c.JSON(http.StatusConflict, gin.H{"error": "Assignment already responded"})
		return
	}

	var input models.RespondAssignmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.assignmentRepo.UpdateStatus(assignment.ID, input.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update assignment"})
		return
	}

	assignment.Status = input.Status
	c.JSON(http.StatusOK, assignment)
}

func (h *AssignmentHandler) GetPendingCount(c *gin.Context) {
	userID := middleware.GetUserID(c)

	count, err := h.assignmentRepo.GetPendingCountByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}
