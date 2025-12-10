package entity

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
)

type Component struct {
	ID       int64           `json:"id"` // 时间戳 ID
	Name     string          `json:"name"`
	Desc     string          `json:"desc"`
	ParentID *int64          `json:"parentId,omitempty"`
	Children []int64         `json:"children,omitempty"`
	Props    json.RawMessage `json:"props,omitempty"`
	Styles   json.RawMessage `json:"styles,omitempty"`
}

// Page 数据库模型 (PostgreSQL JSONB)
type Page struct {
	ID        uint           `gorm:"primaryKey"`
	PageID    string         `gorm:"uniqueIndex;size:64"`
	Schema    datatypes.JSON `gorm:"type:jsonb"`
	Version   int64          `gorm:"default:0"`
	CreatorID string         `gorm:"size:64;index"` // Clerk user_id

	Creator   User `gorm:"foreignKey:CreatorID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
