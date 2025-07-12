package request

type CreateCategoryRequest struct {
	Name        string `json:"name" validate:"required"`
	Icon        string `json:"icon" validate:"required"`
	Description string `json:"description"`
	ParentID    string `json:"parent_id"`
	Status      bool   `json:"status" validate:"required"`
}