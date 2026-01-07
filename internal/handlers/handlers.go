package handlers

import (
	"net/http"
	"productivity-app/internal/dto"
	"productivity-app/internal/services"

	"github.com/gin-gonic/gin"
)

type ProductivityHandler struct {
	service *services.ProductivityService
}

func NewProductivityHandler(service *services.ProductivityService) *ProductivityHandler {
	return &ProductivityHandler{service: service}
}

func (h *ProductivityHandler) GetDashboard(c *gin.Context) {
	resp, err := h.service.GetDashboard(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ProductivityHandler) CreateGoal(c *gin.Context) {
	var req dto.GoalCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateGoal(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

func (h *ProductivityHandler) UpdateGoal(c *gin.Context) {
	id := c.Param("id")
	var req dto.GoalUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateGoal(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *ProductivityHandler) GetGoals(c *gin.Context) {
	resp, err := h.service.GetGoals(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ProductivityHandler) DeleteGoal(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteGoal(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProductivityHandler) Capture(c *gin.Context) {
	var req dto.CaptureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Capture(c.Request.Context(), req.Text); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

func (h *ProductivityHandler) GetActions(c *gin.Context) {
	status := c.Query("status")
	projectID := c.Query("projectId")

	resp, err := h.service.GetActions(c.Request.Context(), status, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ProductivityHandler) CreateAction(c *gin.Context) {
	var req dto.ActionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateAction(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

func (h *ProductivityHandler) UpdateAction(c *gin.Context) {
	id := c.Param("id")
	var req dto.ActionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateAction(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *ProductivityHandler) DeleteAction(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteAction(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProductivityHandler) CompleteAction(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.CompleteAction(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *ProductivityHandler) CreateProject(c *gin.Context) {
	var req dto.ProjectCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateProject(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

func (h *ProductivityHandler) UpdateProject(c *gin.Context) {
	id := c.Param("id")
	var req dto.ProjectUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateProject(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *ProductivityHandler) GetProjects(c *gin.Context) {
	goalID := c.Query("goalId")
	resp, err := h.service.GetProjects(c.Request.Context(), goalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ProductivityHandler) DeleteProject(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteProject(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProductivityHandler) PromoteProject(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.PromoteProject(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *ProductivityHandler) GetWeeklyReview(c *gin.Context) {
	resp, err := h.service.GetWeeklyReview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
