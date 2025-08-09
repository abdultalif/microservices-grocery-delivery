package response

import "github.com/google/uuid"

type CategoryResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string   `json:"name"`
	Icon        string   `json:"icon"`
	Status      string   `json:"status"`
	Slug        string   `json:"slug"`
	Description string   `json:"description"`
	TotalProduct int       `json:"total_products"`
}

type CategoryDetailResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Slug        string `json:"slug"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

type CategoryListHomeResponse struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
	Slug string `json:"slug"`
}

type CategoryListShopResponse struct {
	Name  string                     `json:"name"`
	Slug  string                     `json:"slug"`
	Child []CategoryListShopResponse `json:"child"`
}