package dto

import (
	"productivity-app/internal/models"
	"time"
)

type CaptureRequest struct {
	Text string `json:"text" binding:"required"`
}

type GoalCreateRequest struct {
	Title    string          `json:"title" binding:"required"`
	Priority models.Priority `json:"priority" binding:"required"`
	Active   bool            `json:"active"`
}

type GoalUpdateRequest struct {
	Title    *string          `json:"title"`
	Priority *models.Priority `json:"priority"`
	Active   *bool            `json:"active"`
}

type GoalResponse struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	Priority  models.Priority `json:"priority"`
	Active    bool            `json:"active"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

type ProjectCreateRequest struct {
	GoalID  string               `json:"goalId" binding:"required"`
	Title   string               `json:"title" binding:"required"`
	Outcome string               `json:"outcome" binding:"required"`
	Status  models.ProjectStatus `json:"status" binding:"required"`
}

type ProjectUpdateRequest struct {
	Title   *string               `json:"title"`
	Outcome *string               `json:"outcome"`
	Status  *models.ProjectStatus `json:"status"`
}

type ProjectResponse struct {
	ID        string               `json:"id"`
	GoalID    string               `json:"goalId"`
	Title     string               `json:"title"`
	Outcome   string               `json:"outcome"`
	Status    models.ProjectStatus `json:"status"`
	CreatedAt time.Time            `json:"createdAt"`
	UpdatedAt time.Time            `json:"updatedAt"`
}

type ActionCreateRequest struct {
	ProjectID   string               `json:"projectId" binding:"required"`
	Description string               `json:"description" binding:"required"`
	Context     models.ActionContext `json:"context"`
	Energy      models.Energy        `json:"energy"`
	Status      models.ActionStatus  `json:"status"`
}

type ActionUpdateRequest struct {
	Description *string               `json:"description"`
	Context     *models.ActionContext `json:"context"`
	Energy      *models.Energy        `json:"energy"`
	Status      *models.ActionStatus  `json:"status"`
}

type NextActionResponse struct {
	ID          string               `json:"id"`
	ProjectID   string               `json:"projectId"`
	Description string               `json:"description"`
	Context     models.ActionContext `json:"context"`
	Energy      models.Energy        `json:"energy"`
	Status      models.ActionStatus  `json:"status"`
	CreatedAt   time.Time            `json:"createdAt"`
	UpdatedAt   time.Time            `json:"updatedAt"`
}

type DashboardResponse struct {
	NextActions []NextActionResponse `json:"nextActions"`
}

type WeeklyReviewResponse struct {
	ProjectsWithoutCurrentAction []ProjectResponse    `json:"projectsWithoutCurrentAction"`
	CompletedActions             []NextActionResponse `json:"completedActions"`
	BacklogProjects              []ProjectResponse    `json:"backlogProjects"`
}
