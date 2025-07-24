package request

type ProductRequest struct {
	ProductName        string                `json:"name" validate:"required"`
	CategorySlug       string                `json:"category_slug" validate:"required"`
	Unit               string                `json:"unit" validate:"required"`
	Variant            string                `json:"variant" validate:"required"`
	ProductDescription string                `json:"description" validate:"required"`
	Status             string                `json:"status" validate:"required"`
	VariantDetail      []ProductDetailRequst `json:"variant_detail" validate:"required"`
}

type ProductDetailRequst struct {
	Stock        int    `json:"stock" validate:"required,gt=0"`
	ProductImage string `json:"product_image" validate:"required"`
	Weight       int    `json:"weight" validate:"required,gt=0"`
	SalePrice    int64  `json:"sale_price" validate:"required,gt=0"`
	RegulerPrice int64  `json:"reguler_price" validate:"required,gt=0"`
}
