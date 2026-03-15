package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

// Plan represents a subscription plan that users can subscribe to.
type Plan struct {
	ID          uuid.UUID
	Name        string
	Slug        valueobject.Slug
	Description string
	Price       float64
	Level       int
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewPlan creates a new Plan entity.
func NewPlan(name string, slug valueobject.Slug, description string, price float64, level int) *Plan {
	now := time.Now().UTC()
	return &Plan{
		ID:          uuid.New(),
		Name:        name,
		Slug:        slug,
		Description: description,
		Price:       price,
		Level:       level,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Deactivate marks the plan as inactive.
func (p *Plan) Deactivate() {
	p.IsActive = false
	p.UpdatedAt = time.Now().UTC()
}

// Activate marks the plan as active.
func (p *Plan) Activate() {
	p.IsActive = true
	p.UpdatedAt = time.Now().UTC()
}
