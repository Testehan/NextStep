package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Priority string

const (
	PriorityLow    Priority = "LOW"
	PriorityMedium Priority = "MEDIUM"
	PriorityHigh   Priority = "HIGH"
)

type ProjectStatus string

const (
	ProjectStatusActive  ProjectStatus = "ACTIVE"
	ProjectStatusBacklog ProjectStatus = "BACKLOG"
	ProjectStatusDone    ProjectStatus = "DONE"
)

type ActionContext string

const (
	ContextDeepWork ActionContext = "DEEP_WORK"
	ContextQuick    ActionContext = "QUICK"
	ContextPhone    ActionContext = "PHONE"
	ContextErrands  ActionContext = "ERRANDS"
)

type Energy string

const (
	EnergyHigh Energy = "HIGH"
	EnergyLow  Energy = "LOW"
)

type ActionStatus string

const (
	ActionStatusCurrent ActionStatus = "CURRENT"
	ActionStatusQueued  ActionStatus = "QUEUED"
	ActionStatusDone    ActionStatus = "DONE"
)

type Goal struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string        `bson:"title" json:"title"`
	Priority  Priority      `bson:"priority" json:"priority"`
	Active    bool          `bson:"active" json:"active"`
	CreatedAt time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time     `bson:"updatedAt" json:"updatedAt"`
}

type Project struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	GoalID    bson.ObjectID `bson:"goalId" json:"goalId"`
	Title     string        `bson:"title" json:"title"`
	Outcome   string        `bson:"outcome" json:"outcome"`
	Status    ProjectStatus `bson:"status" json:"status"`
	CreatedAt time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time     `bson:"updatedAt" json:"updatedAt"`
}

type NextAction struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID   bson.ObjectID `bson:"projectId" json:"projectId"`
	Description string        `bson:"description" json:"description"`
	Context     ActionContext `bson:"context" json:"context"`
	Energy      Energy        `bson:"energy" json:"energy"`
	Status      ActionStatus  `bson:"status" json:"status"`
	CreatedAt   time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time     `bson:"updatedAt" json:"updatedAt"`
}
