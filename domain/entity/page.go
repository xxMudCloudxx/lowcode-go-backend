package entity

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
)

// ================= Schema 结构定义 =================

// PageSchema 页面 Schema 的完整结构
// Schema 是存储在数据库 JSONB 字段中的数据结构
type PageSchema struct {
	RootID     int64                `json:"rootId"`
	Components map[string]Component `json:"components"`
}

// Component 组件结构
type Component struct {
	ID       int64           `json:"id"` // 时间戳 ID
	Name     string          `json:"name"`
	Desc     string          `json:"desc"`
	ParentID *int64          `json:"parentId,omitempty"`
	Children []int64         `json:"children,omitempty"`
	Props    json.RawMessage `json:"props,omitempty"`
	Styles   json.RawMessage `json:"styles,omitempty"`
}

// NewDefaultSchema 创建默认的空白 Schema
// ⚠️ 使用强类型结构体初始化，避免硬编码 JSON 字符串
func NewDefaultSchema() *PageSchema {
	rootID := int64(1)
	return &PageSchema{
		RootID: rootID,
		Components: map[string]Component{
			"1": {
				ID:       rootID,
				Name:     "Page",
				Desc:     "页面根节点",
				ParentID: nil,
				Children: []int64{},
				Props:    json.RawMessage(`{}`),
				Styles:   json.RawMessage(`{}`),
			},
		},
	}
}

// MarshalJSON 将 Schema 序列化为 JSON bytes
func (s *PageSchema) MarshalJSON() ([]byte, error) {
	type Alias PageSchema
	return json.Marshal((*Alias)(s))
}

// ToBytes 将 Schema 序列化为 []byte，便于存储
func (s *PageSchema) ToBytes() ([]byte, error) {
	return json.Marshal(s)
}

// ================= Page 数据库模型 =================

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
