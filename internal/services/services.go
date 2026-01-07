package services

import (
	"context"
	"errors"
	"productivity-app/internal/dto"
	"productivity-app/internal/models"
	"productivity-app/internal/repositories"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ProductivityService struct {
	goalRepo    *repositories.GoalRepository
	projectRepo *repositories.ProjectRepository
	actionRepo  *repositories.ActionRepository
}

func NewProductivityService(
	goalRepo *repositories.GoalRepository,
	projectRepo *repositories.ProjectRepository,
	actionRepo *repositories.ActionRepository,
) *ProductivityService {
	return &ProductivityService{
		goalRepo:    goalRepo,
		projectRepo: projectRepo,
		actionRepo:  actionRepo,
	}
}

func (s *ProductivityService) GetDashboard(ctx context.Context) (*dto.DashboardResponse, error) {
	activeProjects, err := s.projectRepo.Find(ctx, bson.M{"status": models.ProjectStatusActive})
	if err != nil {
		return nil, err
	}

	projectIDs := make([]bson.ObjectID, len(activeProjects))
	for i, p := range activeProjects {
		projectIDs[i] = p.ID
	}

	actions, err := s.actionRepo.Find(ctx, bson.M{
		"projectId": bson.M{"$in": projectIDs},
		"status":    models.ActionStatusCurrent,
	}, options.Find().SetLimit(5))
	if err != nil {
		return nil, err
	}

	resp := &dto.DashboardResponse{NextActions: []dto.NextActionResponse{}}
	for _, a := range actions {
		resp.NextActions = append(resp.NextActions, dto.NextActionResponse{
			ID:          a.ID.Hex(),
			ProjectID:   a.ProjectID.Hex(),
			Description: a.Description,
			Context:     a.Context,
			Energy:      a.Energy,
			Status:      a.Status,
		})
	}

	return resp, nil
}

func (s *ProductivityService) CreateGoal(ctx context.Context, req dto.GoalCreateRequest) error {
	goal := &models.Goal{
		Title:    req.Title,
		Priority: req.Priority,
		Active:   req.Active,
	}
	return s.goalRepo.Create(ctx, goal)
}

func (s *ProductivityService) GetGoals(ctx context.Context) ([]dto.GoalResponse, error) {
	goals, err := s.goalRepo.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	resp := []dto.GoalResponse{}
	for _, g := range goals {
		resp = append(resp, dto.GoalResponse{
			ID:       g.ID.Hex(),
			Title:    g.Title,
			Priority: g.Priority,
			Active:   g.Active,
		})
	}
	return resp, nil
}

func (s *ProductivityService) Capture(ctx context.Context, text string) error {
	// Simple parsing: short, specific -> NextAction, vague -> Project
	// This is a placeholder for more sophisticated logic
	isVague := len(strings.Split(text, " ")) < 3 || strings.Contains(strings.ToLower(text), "maybe")

	if isVague {
		// Create a Project (mocking a default GoalID for now, or could be a "Inbox" Goal)
		project := &models.Project{
			Title:  text,
			Status: models.ProjectStatusBacklog,
		}
		return s.projectRepo.Create(ctx, project)
	} else {
		// Create a NextAction (mocking a default ProjectID for now, or could be "Inbox" Project)
		action := &models.NextAction{
			Description: text,
			Status:      models.ActionStatusQueued,
			Context:     models.ContextQuick,
			Energy:      models.EnergyLow,
		}
		return s.actionRepo.Create(ctx, action)
	}
}

func (s *ProductivityService) CreateAction(ctx context.Context, req dto.ActionCreateRequest) error {
	projectID, _ := bson.ObjectIDFromHex(req.ProjectID)

	// Business Rule: Each Project has max ONE NextAction with status = CURRENT
	if req.Status == models.ActionStatusCurrent {
		existing, err := s.actionRepo.FindOne(ctx, bson.M{
			"projectId": projectID,
			"status":    models.ActionStatusCurrent,
		})
		if err == nil && existing != nil {
			return errors.New("each project has max one NextAction with status = CURRENT")
		}
	}

	action := &models.NextAction{
		ProjectID:   projectID,
		Description: req.Description,
		Context:     req.Context,
		Energy:      req.Energy,
		Status:      req.Status,
	}

	return s.actionRepo.Create(ctx, action)
}

func (s *ProductivityService) UpdateAction(ctx context.Context, actionID string, req dto.ActionUpdateRequest) error {
	id, err := bson.ObjectIDFromHex(actionID)
	if err != nil {
		return err
	}

	action, err := s.actionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Description != nil {
		action.Description = *req.Description
	}
	if req.Context != nil {
		action.Context = *req.Context
	}
	if req.Energy != nil {
		action.Energy = *req.Energy
	}
	if req.Status != nil {
		// Business Rule: Each Project has max ONE NextAction with status = CURRENT
		if *req.Status == models.ActionStatusCurrent && action.Status != models.ActionStatusCurrent {
			existing, err := s.actionRepo.FindOne(ctx, bson.M{
				"projectId": action.ProjectID,
				"status":    models.ActionStatusCurrent,
			})
			if err == nil && existing != nil {
				return errors.New("each project has max one NextAction with status = CURRENT")
			}
		}
		action.Status = *req.Status
	}

	return s.actionRepo.Update(ctx, action)
}

func (s *ProductivityService) CompleteAction(ctx context.Context, actionID string) error {
	id, err := bson.ObjectIDFromHex(actionID)
	if err != nil {
		return err
	}

	action, err := s.actionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	action.Status = models.ActionStatusDone
	if err := s.actionRepo.Update(ctx, action); err != nil {
		return err
	}

	// Promote next QUEUED action to CURRENT
	nextAction, err := s.actionRepo.FindOne(ctx, bson.M{
		"projectId": action.ProjectID,
		"status":    models.ActionStatusQueued,
	})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil // No more actions to promote
		}
		return err
	}

	nextAction.Status = models.ActionStatusCurrent
	return s.actionRepo.Update(ctx, nextAction)
}

func (s *ProductivityService) CreateProject(ctx context.Context, req dto.ProjectCreateRequest) error {
	goalID, _ := bson.ObjectIDFromHex(req.GoalID)
	project := &models.Project{
		GoalID:  goalID,
		Title:   req.Title,
		Outcome: req.Outcome,
		Status:  req.Status,
	}

	if project.Status == models.ProjectStatusActive {
		count, err := s.projectRepo.CountActive(ctx)
		if err != nil {
			return err
		}
		if count >= 3 {
			return errors.New("max 3 active projects allowed")
		}
	}

	return s.projectRepo.Create(ctx, project)
}

func (s *ProductivityService) GetProjects(ctx context.Context) ([]dto.ProjectResponse, error) {
	projects, err := s.projectRepo.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	resp := []dto.ProjectResponse{}
	for _, p := range projects {
		resp = append(resp, dto.ProjectResponse{
			ID:      p.ID.Hex(),
			GoalID:  p.GoalID.Hex(),
			Title:   p.Title,
			Outcome: p.Outcome,
			Status:  p.Status,
		})
	}
	return resp, nil
}

func (s *ProductivityService) PromoteProject(ctx context.Context, projectID string) error {
	id, err := bson.ObjectIDFromHex(projectID)
	if err != nil {
		return err
	}

	count, err := s.projectRepo.CountActive(ctx)
	if err != nil {
		return err
	}
	if count >= 3 {
		return errors.New("max 3 active projects allowed")
	}

	return s.projectRepo.UpdateStatus(ctx, id, models.ProjectStatusActive)
}

func (s *ProductivityService) GetWeeklyReview(ctx context.Context) (*dto.WeeklyReviewResponse, error) {
	// projects without CURRENT action
	allProjects, err := s.projectRepo.Find(ctx, bson.M{"status": bson.M{"$ne": models.ProjectStatusDone}})
	if err != nil {
		return nil, err
	}

	projectsWithoutCurrent := []dto.ProjectResponse{}
	for _, p := range allProjects {
		_, err := s.actionRepo.FindOne(ctx, bson.M{
			"projectId": p.ID,
			"status":    models.ActionStatusCurrent,
		})
		if err != nil && errors.Is(err, mongo.ErrNoDocuments) {
			projectsWithoutCurrent = append(projectsWithoutCurrent, dto.ProjectResponse{
				ID:      p.ID.Hex(),
				GoalID:  p.GoalID.Hex(),
				Title:   p.Title,
				Outcome: p.Outcome,
				Status:  p.Status,
			})
		}
	}

	// completed actions
	doneActions, err := s.actionRepo.Find(ctx, bson.M{"status": models.ActionStatusDone})
	if err != nil {
		return nil, err
	}
	completedActions := []dto.NextActionResponse{}
	for _, a := range doneActions {
		completedActions = append(completedActions, dto.NextActionResponse{
			ID:          a.ID.Hex(),
			ProjectID:   a.ProjectID.Hex(),
			Description: a.Description,
			Status:      a.Status,
		})
	}

	// backlog projects
	backlogProjectsData, err := s.projectRepo.Find(ctx, bson.M{"status": models.ProjectStatusBacklog})
	if err != nil {
		return nil, err
	}
	backlogProjects := []dto.ProjectResponse{}
	for _, p := range backlogProjectsData {
		backlogProjects = append(backlogProjects, dto.ProjectResponse{
			ID:      p.ID.Hex(),
			GoalID:  p.GoalID.Hex(),
			Title:   p.Title,
			Outcome: p.Outcome,
			Status:  p.Status,
		})
	}

	return &dto.WeeklyReviewResponse{
		ProjectsWithoutCurrentAction: projectsWithoutCurrent,
		CompletedActions:             completedActions,
		BacklogProjects:              backlogProjects,
	}, nil
}
