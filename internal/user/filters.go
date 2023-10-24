package user

import (
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
