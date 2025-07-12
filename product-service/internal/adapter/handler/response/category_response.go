package response

import "github.com/google/uuid"

type CategoryResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string   `json:"name"`
	Icon        string   `json:"icon"`
	Status      bool   `json:"status"`
	Slug        string   `json:"slug"`
	Description string   `json:"description"`
	TotalProduct int       `json:"total_products"`
}



type CategoryDetailResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Slug        string `json:"slug"`
	Status      bool `json:"status"`
	Description string `json:"description"`
}