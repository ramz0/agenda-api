package handlers

import (
	"net/http"

	"agenda-api/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userRepo *repository.UserRepository
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

func (h *UserHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if len(query) < 2 {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	users, err := h.userRepo.Search(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error searching users"})
		return
	}

	// Convert to response (exclude sensitive data)
	var response []gin.H
	for _, user := range users {
		response = append(response, gin.H{
			"id":        user.ID,
			"name":      user.Name,
			"email":     user.Email,
			"role":      user.Role,
			"createdAt": user.CreatedAt,
		})
	}

	if response == nil {
		response = []gin.H{}
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.userRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching users"})
		return
	}

	var response []gin.H
	for _, user := range users {
		response = append(response, gin.H{
			"id":        user.ID,
			"name":      user.Name,
			"email":     user.Email,
			"role":      user.Role,
			"createdAt": user.CreatedAt,
		})
	}

	if response == nil {
		response = []gin.H{}
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}
