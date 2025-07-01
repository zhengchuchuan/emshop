package db

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        int32     `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"column:add_time"`
	UpdatedAt time.Time `gorm:"column:update_time"`
	DeletedAt gorm.DeletedAt	`gorm:"column:delete_time"`
	IsDeleted bool		`gorm:"column:is_deleted"` //逻辑删除
}