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
			CreatedAt:   a.CreatedAt,
			UpdatedAt:   a.UpdatedAt,
		})
	}

	return resp, nil
}

func (s *ProductivityService) CreateGoal(ctx context.Context, req dto.GoalCreateRequest) error {
	now := time.Now()
	goal := &models.Goal{
		Title:     req.Title,
		Priority:  req.Priority,
		Active:    req.Active,
		CreatedAt: now,
		UpdatedAt: now,
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
			ID:        g.ID.Hex(),
			Title:     g.Title,
			Priority:  g.Priority,
			Active:    g.Active,
			CreatedAt: g.CreatedAt,
			UpdatedAt: g.UpdatedAt,
		})
	}
	return resp, nil
}

func (s *ProductivityService) Capture(ctx context.Context, text string) error {
	// Simple parsing: short, specific -> NextAction, vague -> Project
	// This is a placeholder for more sophisticated logic
	isVague := len(strings.Split(text, " ")) < 3 || strings.Contains(strings.ToLower(text), "maybe")
	now := time.Now()

	if isVague {
		// Create a Project (mocking a default GoalID for now, or could be a "Inbox" Goal)
		project := &models.Project{
			Title:     text,
			Status:    models.ProjectStatusBacklog,
			CreatedAt: now,
			UpdatedAt: now,
		}
		return s.projectRepo.Create(ctx, project)
	} else {
		// Create a NextAction (mocking a default ProjectID for now, or could be "Inbox" Project)
		action := &models.NextAction{
			Description: text,
			Status:      models.ActionStatusQueued,
			Context:     models.ContextQuick,
			Energy:      models.EnergyLow,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		return s.actionRepo.Create(ctx, action)
	}
}

func (s *ProductivityService) GetActions(ctx context.Context, status string, projectID string) ([]dto.NextActionResponse, error) {
	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}
	if projectID != "" {
		pID, err := bson.ObjectIDFromHex(projectID)
		if err == nil {
			filter["projectId"] = pID
		}
	}

	actions, err := s.actionRepo.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	resp := []dto.NextActionResponse{}
	for _, a := range actions {
		resp = append(resp, dto.NextActionResponse{
			ID:          a.ID.Hex(),
			ProjectID:   a.ProjectID.Hex(),
			Description: a.Description,
			Context:     a.Context,
			Energy:      a.Energy,
			Status:      a.Status,
			CreatedAt:   a.CreatedAt,
			UpdatedAt:   a.UpdatedAt,
		})
	}
	return resp, nil
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

	now := time.Now()
	action := &models.NextAction{
		ProjectID:   projectID,
		Description: req.Description,
		Context:     req.Context,
		Energy:      req.Energy,
		Status:      req.Status,
		CreatedAt:   now,
		UpdatedAt:   now,
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

	action.UpdatedAt = time.Now()
	if err := s.actionRepo.Update(ctx, action); err != nil {
		return err
	}

	// Trigger promotion if status was changed to DONE
	if req.Status != nil && *req.Status == models.ActionStatusDone {
		_ = s.promoteNextQueued(ctx, action.ProjectID)
	}

	return nil
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

	now := time.Now()
	action.Status = models.ActionStatusDone
	action.UpdatedAt = now
	if err := s.actionRepo.Update(ctx, action); err != nil {
		return err
	}

	return s.promoteNextQueued(ctx, action.ProjectID)
}

func (s *ProductivityService) promoteNextQueued(ctx context.Context, projectID bson.ObjectID) error {
	// Only promote if there is NO current action
	current, err := s.actionRepo.FindOne(ctx, bson.M{
		"projectId": projectID,
		"status":    models.ActionStatusCurrent,
	})
	if err == nil && current != nil {
		return nil
	}

	// Find the OLDEST queued action (sort by _id)
	nextAction, err := s.actionRepo.FindOne(ctx, bson.M{
		"projectId": projectID,
		"status":    models.ActionStatusQueued,
	}, options.FindOne().SetSort(bson.M{"_id": 1}))

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil
		}
		return err
	}

	nextAction.Status = models.ActionStatusCurrent
	nextAction.UpdatedAt = time.Now()
	return s.actionRepo.Update(ctx, nextAction)
}

func (s *ProductivityService) CreateProject(ctx context.Context, req dto.ProjectCreateRequest) error {
	goalID, _ := bson.ObjectIDFromHex(req.GoalID)
	now := time.Now()
	project := &models.Project{
		GoalID:    goalID,
		Title:     req.Title,
		Outcome:   req.Outcome,
		Status:    req.Status,
		CreatedAt: now,
		UpdatedAt: now,
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

func (s *ProductivityService) GetProjects(ctx context.Context, goalID string) ([]dto.ProjectResponse, error) {
	filter := bson.M{}
	if goalID != "" {
		id, err := bson.ObjectIDFromHex(goalID)
		if err == nil {
			filter["goalId"] = id
		}
	}

	projects, err := s.projectRepo.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	resp := []dto.ProjectResponse{}
	for _, p := range projects {
		resp = append(resp, dto.ProjectResponse{
			ID:        p.ID.Hex(),
			GoalID:    p.GoalID.Hex(),
			Title:     p.Title,
			Outcome:   p.Outcome,
			Status:    p.Status,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
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

	// Update with timestamp
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	project.Status = models.ProjectStatusActive
	project.UpdatedAt = time.Now()
	
	return s.projectRepo.Update(ctx, project)
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
				ID:        p.ID.Hex(),
				GoalID:    p.GoalID.Hex(),
				Title:     p.Title,
				Outcome:   p.Outcome,
				Status:    p.Status,
				CreatedAt: p.CreatedAt,
				UpdatedAt: p.UpdatedAt,
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
			CreatedAt:   a.CreatedAt,
			UpdatedAt:   a.UpdatedAt,
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
			ID:        p.ID.Hex(),
			GoalID:    p.GoalID.Hex(),
			Title:     p.Title,
			Outcome:   p.Outcome,
			Status:    p.Status,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		})
	}

	return &dto.WeeklyReviewResponse{
		ProjectsWithoutCurrentAction: projectsWithoutCurrent,
		CompletedActions:             completedActions,
		BacklogProjects:              backlogProjects,
	}, nil
}
