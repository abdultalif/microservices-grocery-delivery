package response

import "github.com/google/uuid"

type CartResponse struct {
	ID            uuid.UUID `json:"id"`
	ProductName   string    `json:"product_name"`
	ProductImage  string    `json:"product_image"`
	ProductStatus string    `json:"product_status"`
	SalePrice     int64     `json:"sale_price"`
	Quantity      int64     `json:"quantity"`
	Unit          string    `json:"unit"`
	Weight        int64     `json:"weight"`
}
