package repositories

import (
	"context"
	"productivity-app/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type GoalRepository struct {
	collection *mongo.Collection
}

func NewGoalRepository(db *mongo.Database) *GoalRepository {
	return &GoalRepository{
		collection: db.Collection("goals"),
	}
}

func (r *GoalRepository) Create(ctx context.Context, goal *models.Goal) error {
	res, err := r.collection.InsertOne(ctx, goal)
	if err != nil {
		return err
	}
	goal.ID = res.InsertedID.(bson.ObjectID)
	return nil
}

func (r *GoalRepository) Find(ctx context.Context, filter bson.M) ([]models.Goal, error) {
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var goals []models.Goal
	if err = cursor.All(ctx, &goals); err != nil {
		return nil, err
	}
	return goals, nil
}

type ProjectRepository struct {
	collection *mongo.Collection
}

func NewProjectRepository(db *mongo.Database) *ProjectRepository {
	repo := &ProjectRepository{
		collection: db.Collection("projects"),
	}
	repo.initIndexes(context.Background())
	return repo
}

func (r *ProjectRepository) initIndexes(ctx context.Context) {
	_, _ = r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "status", Value: 1}},
	})
}

func (r *ProjectRepository) Create(ctx context.Context, project *models.Project) error {
	res, err := r.collection.InsertOne(ctx, project)
	if err != nil {
		return err
	}
	project.ID = res.InsertedID.(bson.ObjectID)
	return nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id bson.ObjectID) (*models.Project, error) {
	var project models.Project
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&project)
	return &project, err
}

func (r *ProjectRepository) UpdateStatus(ctx context.Context, id bson.ObjectID, status models.ProjectStatus) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"status": status}})
	return err
}

func (r *ProjectRepository) Update(ctx context.Context, project *models.Project) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": project.ID}, bson.M{"$set": project})
	return err
}

func (r *ProjectRepository) Find(ctx context.Context, filter bson.M) ([]models.Project, error) {
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var projects []models.Project
	if err = cursor.All(ctx, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *ProjectRepository) CountActive(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"status": models.ProjectStatusActive})
}

type ActionRepository struct {
	collection *mongo.Collection
}

func NewActionRepository(db *mongo.Database) *ActionRepository {
	repo := &ActionRepository{
		collection: db.Collection("actions"),
	}
	repo.initIndexes(context.Background())
	return repo
}

func (r *ActionRepository) initIndexes(ctx context.Context) {
	_, _ = r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "projectId", Value: 1},
			{Key: "status", Value: 1},
		},
	})
}

func (r *ActionRepository) Create(ctx context.Context, action *models.NextAction) error {
	res, err := r.collection.InsertOne(ctx, action)
	if err != nil {
		return err
	}
	action.ID = res.InsertedID.(bson.ObjectID)
	return nil
}

func (r *ActionRepository) GetByID(ctx context.Context, id bson.ObjectID) (*models.NextAction, error) {
	var action models.NextAction
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&action)
	return &action, err
}

func (r *ActionRepository) Update(ctx context.Context, action *models.NextAction) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": action.ID}, bson.M{"$set": action})
	return err
}

func (r *ActionRepository) Find(ctx context.Context, filter bson.M, opts ...options.Lister[options.FindOptions]) ([]models.NextAction, error) {
	cursor, err := r.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	var actions []models.NextAction
	if err = cursor.All(ctx, &actions); err != nil {
		return nil, err
	}
	return actions, nil
}

func (r *ActionRepository) FindOne(ctx context.Context, filter bson.M, opts ...options.Lister[options.FindOneOptions]) (*models.NextAction, error) {
	var action models.NextAction
	err := r.collection.FindOne(ctx, filter, opts...).Decode(&action)
	if err != nil {
		return nil, err
	}
	return &action, nil
}
