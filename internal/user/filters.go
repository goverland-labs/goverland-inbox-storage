package user

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Filter interface {
	Apply(*gorm.DB) *gorm.DB
}

type PageFilter struct {
	Offset int
	Limit  int
}

func (f PageFilter) Apply(db *gorm.DB) *gorm.DB {
	return db.Offset(f.Offset).Limit(f.Limit)
}

type UserIDFilter struct {
	ID uuid.UUID
}

func (f UserIDFilter) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("user_id = ?", f.ID)
}

type OrderByFilter struct {
	Field string
	Desc  bool
}

func (f OrderByFilter) Apply(db *gorm.DB) *gorm.DB {
	return db.Order(clause.OrderByColumn{
		Column: clause.Column{Name: f.Field},
		Desc:   true,
	})
}

type GroupBy struct {
	Field string
}

func (f GroupBy) Apply(db *gorm.DB) *gorm.DB {
	return db.Distinct(f.Field)
}

type TypeFilter struct {
	Type RecentlyType
}

func (f TypeFilter) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("type = ?", f.Type)
}

type ActivityFilterBetween struct {
	From time.Time
	To   time.Time
}

func (f ActivityFilterBetween) Apply(db *gorm.DB) *gorm.DB {
	var (
		dummy = Activity{}
		_     = dummy.CreatedAt
	)

	if !f.From.IsZero() {
		db = db.Where("created_at >= ?", f.From)
	}

	if !f.To.IsZero() {
		db = db.Where("created_at <= ?", f.To)
	}

	return db
}

type ActivityFilterUserID struct {
	UserID uuid.UUID
}

func (f ActivityFilterUserID) Apply(db *gorm.DB) *gorm.DB {
	var (
		dummy = Activity{}
		_     = dummy.UserID
	)

	return db.Where("user_id = ?", f.UserID)
}

type ActivityFilterUserIDOrderBy struct {
	Field     string
	Direction string
}

func (f ActivityFilterUserIDOrderBy) Apply(db *gorm.DB) *gorm.DB {
	return db.Order(fmt.Sprintf("%s %s", f.Field, f.Direction))
}
