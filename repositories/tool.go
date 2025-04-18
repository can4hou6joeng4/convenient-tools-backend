package repositories

import (
	"github.com/can4hou6joeng4/convenient-tools-project-v1-backend/models"
	"gorm.io/gorm"
)

type ToolRepository struct {
	db *gorm.DB
}

func (r *ToolRepository) GetAllTools() ([]*models.Tool, error) {
	tools := []*models.Tool{}
	if err := r.db.Model(&models.Tool{}).Preload("Steps").Preload("Categories").Find(&tools).Error; err != nil {
		return nil, err
	}
	return tools, nil
}
func (r *ToolRepository) CreateTool(tool *models.Tool) error {
	return r.db.Create(tool).Error
}
func NewToolRepository(db *gorm.DB) *ToolRepository {
	return &ToolRepository{
		db: db,
	}
}
