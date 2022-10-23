package repository

import "github.com/google/uuid"

type Site struct {
	ID       uuid.UUID
	Url      string `gorm:"unique;not null"`
	Category string
	Status   string
	ParentID *uuid.UUID
	Parent   *Site `gorm:"foreignkey:ParentID"`
}
