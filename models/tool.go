package models

// 分类结构
type Category struct {
	Base
	UUID string `json:"id" gorm:"type:uuid;default:gen_random_uuid();uniqueIndex"`
	Name string `json:"name" gorm:"size:50;not null"`
	// 多对多关系
	Tools []*Tool `json:"-" gorm:"many2many:tool_categories;"`
}

// 工具步骤结构
type Step struct {
	Base
	Title  string `json:"title" gorm:"size:100;not null"`
	Desc   string `json:"desc" gorm:"size:255;not null"`
	ToolID uint   `json:"-" gorm:"index"` // 外键
	Order  int    `json:"order" gorm:"not null;default:0"`
}

// 工具结构
type Tool struct {
	Base
	UUID        string      `json:"id" gorm:"type:uuid;default:gen_random_uuid();uniqueIndex"`
	Name        string      `json:"name" gorm:"size:100;not null"`
	Description string      `json:"description" gorm:"size:255"`
	Icon        string      `json:"icon" gorm:"size:50"`
	Steps       []*Step     `json:"steps" gorm:"foreignKey:ToolID;constraint:OnDelete:CASCADE"`
	Categories  []*Category `json:"-" gorm:"many2many:tool_categories;"`
}

// ToolCategory 多对多中间表
type ToolCategory struct {
	ToolID     uint `gorm:"primaryKey"`
	CategoryID uint `gorm:"primaryKey"`
}

// 工具包含的分类ID列表 - 用于API
func (t *Tool) CategoryIDs() []string {
	var ids []string
	for _, category := range t.Categories {
		ids = append(ids, category.UUID)
	}
	return ids
}

// 工具API响应格式
func (t *Tool) ToResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":          t.UUID,
		"name":        t.Name,
		"description": t.Description,
		"icon":        t.Icon,
		"categories":  t.CategoryIDs(),
		"steps":       t.Steps,
	}
}

// 工具列表响应结构 - 仅用于API响应
type ToolsResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Categories []Category `json:"categories"`
		Tools      []Tool     `json:"tools"`
	} `json:"data"`
}

// 工具仓库接口
type NewToolRepository interface {
	GetAllTools() ([]*Tool, error)
}
