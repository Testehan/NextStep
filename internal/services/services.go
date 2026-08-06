package services

import (
	"context"
	"errors"
	"productivity-app/internal/dto"
	"productivity-app/internal/models"
	"productivity-app/internal/repositories"
	"strings"
	"time"

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
	}, options.Find().SetLimit(5).SetSort(bson.D{{Key: "sortOrder", Value: 1}, {Key: "createdAt", Value: 1}}))
	if err != nil {
		return nil, err
	}

	resp := &dto.DashboardResponse{NextActions: []dto.NextActionResponse{}}
	for _, a := range actions {
		projectID := a.ProjectID.Hex()
		if projectID == "000000000000000000000000" {
			projectID = ""
		}
		resp.NextActions = append(resp.NextActions, dto.NextActionResponse{
			ID:          a.ID.Hex(),
			ProjectID:   projectID,
			Description: a.Description,
			Context:     a.Context,
			Energy:      a.Energy,
			Status:      a.Status,
			SortOrder:   a.SortOrder,
			CompletedAt: a.CompletedAt,
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

func (s *ProductivityService) UpdateGoal(ctx context.Context, goalID string, req dto.GoalUpdateRequest) error {
	id, err := bson.ObjectIDFromHex(goalID)
	if err != nil {
		return err
	}

	goal, err := s.goalRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Title != nil {
		goal.Title = *req.Title
	}
	if req.Priority != nil {
		goal.Priority = *req.Priority
	}
	if req.Active != nil {
		goal.Active = *req.Active
	}

	goal.UpdatedAt = time.Now()
	return s.goalRepo.Update(ctx, goal)
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

func (s *ProductivityService) DeleteGoal(ctx context.Context, goalID string) error {
	id, err := bson.ObjectIDFromHex(goalID)
	if err != nil {
		return err
	}

	projects, err := s.projectRepo.Find(ctx, bson.M{"goalId": id})
	if err == nil {
		for _, p := range projects {
			_ = s.actionRepo.DeleteMany(ctx, bson.M{"projectId": p.ID})
		}
	}
	_ = s.projectRepo.DeleteMany(ctx, bson.M{"goalId": id})

	return s.goalRepo.Delete(ctx, id)
}

func (s *ProductivityService) Capture(ctx context.Context, text string) error {
	isVague := len(strings.Split(text, " ")) < 3 || strings.Contains(strings.ToLower(text), "maybe")
	now := time.Now()

	if isVague {
		project := &models.Project{
			Title:     text,
			Status:    models.ProjectStatusBacklog,
			CreatedAt: now,
			UpdatedAt: now,
		}
		return s.projectRepo.Create(ctx, project)
	} else {
		lastAction, err := s.actionRepo.FindOne(ctx, bson.M{}, options.FindOne().SetSort(bson.M{"sortOrder": -1}))
		sortOrder := 0.0
		if err == nil {
			sortOrder = lastAction.SortOrder + 1
		}

		action := &models.NextAction{
			Description: text,
			Status:      models.ActionStatusQueued,
			Context:     models.ContextQuick,
			Energy:      models.EnergyLow,
			SortOrder:   sortOrder,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		return s.actionRepo.Create(ctx, action)
	}
}

func (s *ProductivityService) GetActions(ctx context.Context, status, projectID, from, to, goalID string) ([]dto.NextActionResponse, error) {
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
	if goalID != "" {
		gID, err := bson.ObjectIDFromHex(goalID)
		if err == nil {
			projects, _ := s.projectRepo.Find(ctx, bson.M{"goalId": gID})
			projectIDs := make([]bson.ObjectID, len(projects))
			for i, p := range projects {
				projectIDs[i] = p.ID
			}
			if len(projectIDs) > 0 {
				filter["projectId"] = bson.M{"$in": projectIDs}
			}
		}
	}
	if from != "" {
		fromTime, err := time.Parse(time.RFC3339, from)
		if err == nil {
			field := "createdAt"
			if status == string(models.ActionStatusDone) {
				field = "completedAt"
			}
			filter[field] = bson.M{"$gte": fromTime}
		}
	}
	if to != "" {
		toTime, err := time.Parse(time.RFC3339, to)
		if err == nil {
			field := "createdAt"
			if status == string(models.ActionStatusDone) {
				field = "completedAt"
			}
			if existing, ok := filter[field]; ok {
				filter[field] = bson.M{"$gte": existing.(bson.M)["$gte"], "$lte": toTime}
			} else {
				filter[field] = bson.M{"$lte": toTime}
			}
		}
	}

	actions, err := s.actionRepo.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "sortOrder", Value: 1}, {Key: "createdAt", Value: 1}}))
	if err != nil {
		return nil, err
	}

	resp := []dto.NextActionResponse{}
	for _, a := range actions {
		pID := a.ProjectID.Hex()
		if pID == "000000000000000000000000" {
			pID = ""
		}
		resp = append(resp, dto.NextActionResponse{
			ID:          a.ID.Hex(),
			ProjectID:   pID,
			Description: a.Description,
			Context:     a.Context,
			Energy:      a.Energy,
			Status:      a.Status,
			SortOrder:   a.SortOrder,
			CompletedAt: a.CompletedAt,
			CreatedAt:   a.CreatedAt,
			UpdatedAt:   a.UpdatedAt,
		})
	}
	return resp, nil
}

func (s *ProductivityService) CreateAction(ctx context.Context, req dto.ActionCreateRequest) error {
	var projectID bson.ObjectID
	if req.ProjectID != nil && *req.ProjectID != "" {
		projectID, _ = bson.ObjectIDFromHex(*req.ProjectID)
	}

	if req.Status == models.ActionStatusCurrent && projectID != (bson.ObjectID{}) {
		existing, err := s.actionRepo.FindOne(ctx, bson.M{
			"projectId": projectID,
			"status":    models.ActionStatusCurrent,
		})
		if err == nil && existing != nil {
			return errors.New("each project has max one NextAction with status = CURRENT")
		}
	}

	// Auto-assign sortOrder
	lastAction, err := s.actionRepo.FindOne(ctx, bson.M{"projectId": projectID}, options.FindOne().SetSort(bson.M{"sortOrder": -1}))
	sortOrder := 0.0
	if err == nil {
		sortOrder = lastAction.SortOrder + 1
	}

	now := time.Now()
	action := &models.NextAction{
		ProjectID:   projectID,
		Description: req.Description,
		Context:     req.Context,
		Energy:      req.Energy,
		Status:      req.Status,
		SortOrder:   sortOrder,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Auto-promote if project is ACTIVE and has no CURRENT action
	if action.Status == models.ActionStatusQueued && projectID != (bson.ObjectID{}) {
		project, err := s.projectRepo.GetByID(ctx, projectID)
		if err == nil && project.Status == models.ProjectStatusActive {
			current, _ := s.actionRepo.FindOne(ctx, bson.M{
				"projectId": projectID,
				"status":    models.ActionStatusCurrent,
			})
			if current == nil {
				action.Status = models.ActionStatusCurrent
			}
		}
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

	newStatus := action.Status
	if req.Status != nil {
		newStatus = *req.Status
	}

	newProjectID := action.ProjectID
	if req.ProjectID != nil {
		if *req.ProjectID == "" {
			newProjectID = bson.ObjectID{}
		} else {
			pID, err := bson.ObjectIDFromHex(*req.ProjectID)
			if err != nil {
				return err
			}
			newProjectID = pID
		}
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
	if req.SortOrder != nil {
		action.SortOrder = *req.SortOrder
	}

	if newStatus == models.ActionStatusCurrent && (action.Status != models.ActionStatusCurrent || action.ProjectID != newProjectID) {
		if newProjectID != (bson.ObjectID{}) {
			existing, err := s.actionRepo.FindOne(ctx, bson.M{
				"projectId": newProjectID,
				"status":    models.ActionStatusCurrent,
				"_id":       bson.M{"$ne": action.ID},
			})
			if err == nil && existing != nil {
				return errors.New("each project has max one NextAction with status = CURRENT")
			}
		}
	}

	action.Status = newStatus
	action.ProjectID = newProjectID

	if req.Status != nil && *req.Status == models.ActionStatusDone {
		now := time.Now()
		action.CompletedAt = &now
	} else if req.Status != nil && *req.Status != models.ActionStatusDone {
		action.CompletedAt = nil
	}

	if req.CreatedAt != nil {
		action.CreatedAt = *req.CreatedAt
	}

	action.UpdatedAt = time.Now()
	if err := s.actionRepo.Update(ctx, action); err != nil {
		return err
	}

	if req.Status != nil && *req.Status == models.ActionStatusDone {
		_ = s.promoteNextQueued(ctx, action.ProjectID)
	}

	return nil
}

func (s *ProductivityService) DeleteAction(ctx context.Context, actionID string) error {
	id, err := bson.ObjectIDFromHex(actionID)
	if err != nil {
		return err
	}
	return s.actionRepo.Delete(ctx, id)
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
	action.CompletedAt = &now
	action.UpdatedAt = now
	if err := s.actionRepo.Update(ctx, action); err != nil {
		return err
	}

	return s.promoteNextQueued(ctx, action.ProjectID)
}

func (s *ProductivityService) promoteNextQueued(ctx context.Context, projectID bson.ObjectID) error {
	current, err := s.actionRepo.FindOne(ctx, bson.M{
		"projectId": projectID,
		"status":    models.ActionStatusCurrent,
	})
	if err == nil && current != nil {
		return nil
	}

	nextAction, err := s.actionRepo.FindOne(ctx, bson.M{
		"projectId": projectID,
		"status":    models.ActionStatusQueued,
	}, options.FindOne().SetSort(bson.D{{Key: "sortOrder", Value: 1}, {Key: "createdAt", Value: 1}}))

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

func (s *ProductivityService) UpdateProject(ctx context.Context, projectID string, req dto.ProjectUpdateRequest) error {
	id, err := bson.ObjectIDFromHex(projectID)
	if err != nil {
		return err
	}

	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	oldStatus := project.Status
	if req.GoalID != nil {
		goalID, err := bson.ObjectIDFromHex(*req.GoalID)
		if err == nil {
			project.GoalID = goalID
		}
	}
	if req.Title != nil {
		project.Title = *req.Title
	}
	if req.Outcome != nil {
		project.Outcome = *req.Outcome
	}
	if req.Status != nil {
		if *req.Status == models.ProjectStatusActive && project.Status != models.ProjectStatusActive {
			count, err := s.projectRepo.CountActive(ctx)
			if err != nil {
				return err
			}
			if count >= 3 {
				return errors.New("max 3 active projects allowed")
			}
		}
		project.Status = *req.Status
	}

	project.UpdatedAt = time.Now()
	if err := s.projectRepo.Update(ctx, project); err != nil {
		return err
	}

	// If project became ACTIVE, promote a queued action
	if project.Status == models.ProjectStatusActive && oldStatus != models.ProjectStatusActive {
		_ = s.promoteNextQueued(ctx, project.ID)
	}

	return nil
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

	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	oldStatus := project.Status
	project.Status = models.ProjectStatusActive
	project.UpdatedAt = time.Now()

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return err
	}

	if project.Status == models.ProjectStatusActive && oldStatus != models.ProjectStatusActive {
		_ = s.promoteNextQueued(ctx, project.ID)
	}

	return nil
}

func (s *ProductivityService) DeleteProject(ctx context.Context, projectID string) error {
	id, err := bson.ObjectIDFromHex(projectID)
	if err != nil {
		return err
	}

	_ = s.actionRepo.DeleteMany(ctx, bson.M{"projectId": id})

	return s.projectRepo.Delete(ctx, id)
}

func (s *ProductivityService) GetWeeklyReview(ctx context.Context) (*dto.WeeklyReviewResponse, error) {
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

	doneActions, err := s.actionRepo.Find(ctx, bson.M{"status": models.ActionStatusDone})
	if err != nil {
		return nil, err
	}
	completedActions := []dto.NextActionResponse{}
	for _, a := range doneActions {
		projectID := a.ProjectID.Hex()
		if projectID == "000000000000000000000000" {
			projectID = ""
		}
		completedActions = append(completedActions, dto.NextActionResponse{
			ID:          a.ID.Hex(),
			ProjectID:   projectID,
			Description: a.Description,
			Status:      a.Status,
			SortOrder:   a.SortOrder,
			CompletedAt: a.CompletedAt,
			CreatedAt:   a.CreatedAt,
			UpdatedAt:   a.UpdatedAt,
		})
	}

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
