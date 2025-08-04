package mapper

import (
	"product-service/internal/core/domain/entity"
	"product-service/internal/core/domain/model"
)

func MapModelCategoryToEntity(c model.Category) entity.CategoryEntity {
	return entity.CategoryEntity{
		ID:          c.ID,
		ParentID:    c.ParentID,
		Name:        c.Name,
		Icon:        c.Icon,
		Status:      c.Status,
		Slug:        c.Slug,
		Description: c.Description,
	}
}


func MapModelToEntity(p model.Product) entity.ProductEntity {
	childs := []entity.ProductChildEntity{}
	for _, c := range p.Childs {
		childs = append(childs, entity.ProductChildEntity{
			ID:           c.ID,
			Image:        c.Image,
			Weight:       c.Weight,
			Stock:        c.Stock,
			RegulerPrice: c.RegulerPrice,
			SalePrice:    c.SalePrice,
		})
	}

	return entity.ProductEntity{
		ID:           p.ID,
		CategorySlug: p.CategorySlug,
		ParentID:     p.ParentID,
		Name:         p.Name,
		Image:        p.Image,
		Description:  p.Description,
		RegulerPrice: p.RegulerPrice,
		SalePrice:    p.SalePrice,
		Unit:         p.Unit,
		Weight:       p.Weight,
		Stock:        p.Stock,
		Variant:      p.Variant,
		Status:       p.Status,
		CreatedAt:    p.CreatedAt,
		Child:        childs,
		CategoryName: p.Category.Name,
	}
}
