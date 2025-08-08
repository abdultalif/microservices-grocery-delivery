package response

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