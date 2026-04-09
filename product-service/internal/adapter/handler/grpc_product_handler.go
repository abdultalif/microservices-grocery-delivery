package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errs "github.com/abdultalif/microservices-grocery-delivery/product-service/internal/core/domain/error"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/core/service"
	productPB "github.com/abdultalif/microservices-grocery-delivery/product-service/proto/product"
	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

type GRPCProductHandler struct {
	productPB.UnimplementedProductServiceServer
	productService service.ProductServiceInterface
}

func NewGRPCProductHandler(productService service.ProductServiceInterface) *GRPCProductHandler {
	return &GRPCProductHandler{
		productService: productService,
	}
}

func (g *GRPCProductHandler) GetProductByID(ctx context.Context, req *productPB.GetProductRequest) (*productPB.GetProductResponse, error) {
	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		return &productPB.GetProductResponse{
			Success: false,
			Code:    400,
			Message: "Invalid product ID",
			Data:    nil,
		}, nil
	}

	product, err := g.productService.GetByID(ctx, productID)
	if err != nil {
		log.Errorf("[GRPCProductHandler-1] GetProductByID: %v", err)
		if errors.Is(err, errs.ErrProductNotFound) {
			return &productPB.GetProductResponse{
				Success: false,
				Code:    404,
				Message: "Product not found",
				Data:    nil,
			}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if product == nil {
		return &productPB.GetProductResponse{
			Success: false,
			Code:    404,
			Message: "Product not found",
			Data:    nil,
		}, nil
	}

	childProducts := make([]*productPB.ProductChild, len(product.Child))
	for i, child := range product.Child {
		childProducts[i] = &productPB.ProductChild{
			Id:           child.ID.String(),
			Weight:       int32(child.Weight),
			Stock:        int32(child.Stock),
			RegulerPrice: int64(child.RegulerPrice),
			SalePrice:    int64(child.SalePrice),
		}
	}

	parentID := ""
	if product.ParentID != nil && *product.ParentID != uuid.Nil {
		parentID = product.ParentID.String()
	}

	categoryName := ""
	if product.CategoryName != "" {
		categoryName = product.CategoryName
	}

	createdAtStr := ""
	if !product.CreatedAt.IsZero() {
		createdAtStr = product.CreatedAt.Format("2006-01-02 15:04:05")
	}

	return &productPB.GetProductResponse{
		Success: true,
		Code:    200,
		Message: "success",
		Data: &productPB.ProductData{
			Id:           product.ID.String(),
			CategorySlug: product.CategorySlug,
			ParentId:     parentID,
			Name:         product.Name,
			Image:        product.Image,
			Description:  product.Description,
			RegulerPrice: int64(product.RegulerPrice),
			SalePrice:    int64(product.SalePrice),
			Unit:         product.Unit,
			Weight:       int32(product.Weight),
			Stock:        int32(product.Stock),
			Status:       product.Status,
			CategoryName: categoryName,
			CreatedAt:    createdAtStr,
			Child:        childProducts,
		},
	}, nil
}
