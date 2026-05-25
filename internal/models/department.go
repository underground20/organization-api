package models

import (
	"time"
)

type Department struct {
	ID        int         `json:"id" gorm:"primarykey"`
	Name      string      `json:"name" gorm:"not null" validate:"required,max=255"`
	ParentID  *int        `json:"parent_id" gorm:"index"`
	Parent    *Department `json:"-" gorm:"foreignKey:ParentID"`
	CreatedAt time.Time   `json:"created_at"`
}
