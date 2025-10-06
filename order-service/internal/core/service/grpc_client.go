package service

import (
	"context"
	"fmt"
	"time"

	"github.com/abdultalif/microservices-grocery-delivery/order-service/config"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/core/domain/entity"
	errs "github.com/abdultalif/microservices-grocery-delivery/order-service/internal/core/domain/error"
	productPB "github.com/abdultalif/microservices-grocery-delivery/order-service/proto/product"
	userPB "github.com/abdultalif/microservices-grocery-delivery/order-service/proto/user"
	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type GRPCClient struct {
	userClient    userPB.UserServiceClient
	productClient productPB.ProductServiceClient
	userConn      *grpc.ClientConn
	productConn   *grpc.ClientConn
	cfg           *config.Config
}

func (g *GRPCClient) GetUserClient() userPB.UserServiceClient {
	return g.userClient
}

func (g *GRPCClient) GetProductClient() productPB.ProductServiceClient {
	return g.productClient
}

func (g *GRPCClient) GetInternalTokenGRPC() (string, error) {
	req := &userPB.GetInternalTokenRequest{
		ClientId:     g.cfg.App.AuthClientID,
		ClientSecret: g.cfg.App.AuthClientSecret,
	}

	res, err := g.userClient.GetInternalToken(context.Background(), req)
	if err != nil {
		log.Errorf("[GRPCClient-1] GetInternalTokenGRPC: %v", err)
		return "", err
	}

	if !res.Success || res.Data.AccessToken == "" {
		return "", fmt.Errorf("[GRPCClient-2] GetInternalTokenGRPC: failed, msg: %s", res.Message)
	}

	return res.Data.AccessToken, nil
}

func (g *GRPCClient) GetUserByIDGRPC(userID int64, accessToken string) (*entity.CustomerResponseEntity, error) {
	req := &userPB.GetCustomerRequest{
		UserId:      userID,
		AccessToken: accessToken,
	}

	resp, err := g.userClient.GetCustomerByID(context.Background(), req)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errs.ErrNotFoundBuyer
		}
		log.Errorf("[GRPCClient-1] GetUserByIDGRPC: %v", err)
		return nil, err
	}

	if !resp.Success {
		switch resp.Code {
		case 404:
			return nil, errs.ErrNotFoundBuyer
		case 401, 403:
			return nil, fmt.Errorf("user service auth error: %s", resp.Message)
		default:
			return nil, fmt.Errorf("user service error (code %d): %s", resp.Code, resp.Message)
		}
	}

	return &entity.CustomerResponseEntity{
		ID:       resp.Data.Id,
		Name:     resp.Data.Name,
		Email:    resp.Data.Email,
		Phone:    resp.Data.Phone,
		Address:  resp.Data.Address,
		Photo:    resp.Data.Photo,
		Lat:      resp.Data.Lat,
		Lng:      resp.Data.Lng,
		RoleID:   resp.Data.RoleId,
		RoleName: resp.Data.RoleName,
	}, nil
}

// Method untuk GetProductByIDGRPC - pindahkan dari order_service_grpc.go
func (g *GRPCClient) GetProductByIDGRPC(productID string, accessToken string) (*entity.ProductResponseEntity, error) {
	req := &productPB.GetProductRequest{
		ProductId:   productID,
		AccessToken: accessToken,
		IsCustomer:  false,
	}

	resp, err := g.productClient.GetProductByID(context.Background(), req)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errs.ErrNotFoundProduct
		}
		log.Errorf("[GRPCClient-1] GetProductByIDGRPC: %v", err)
		return nil, err
	}

	if !resp.Success {
		return nil, errs.ErrNotFoundProduct
	}

	childEntities := make([]entity.ChildProductResponseEntity, len(resp.Data.Child))
	for i, child := range resp.Data.Child {
		childEntities[i] = entity.ChildProductResponseEntity{
			ID:           uuid.MustParse(child.Id),
			Weight:       int(child.Weight),
			Stock:        int(child.Stock),
			RegulerPrice: float64(child.RegulerPrice),
			SalePrice:    float64(child.SalePrice),
		}
	}

	// PERBAIKAN: Safe UUID parsing dengan error handling
	var productIDParsed uuid.UUID
	var parentIDParsed uuid.UUID

	// Parse Product ID
	if resp.Data.Id != "" {
		productIDParsed, err = uuid.Parse(resp.Data.Id)
		if err != nil {
			log.Errorf("[GRPCClient-2] GetProductByIDGRPC: invalid product ID '%s': %v", resp.Data.Id, err)
			return nil, fmt.Errorf("invalid product ID format")
		}
	} else {
		log.Errorf("[GRPCClient-3] GetProductByIDGRPC: empty product ID")
		return nil, fmt.Errorf("empty product ID")
	}

	// Parse Parent ID (bisa empty)
	if resp.Data.ParentId != "" {
		parentIDParsed, err = uuid.Parse(resp.Data.ParentId)
		if err != nil {
			log.Warnf("[GRPCClient-4] GetProductByIDGRPC: invalid parent ID '%s': %v", resp.Data.ParentId, err)
			parentIDParsed = uuid.Nil // Use nil UUID instead of failing
		}
	} else {
		parentIDParsed = uuid.Nil // Empty parent ID is acceptable
	}

	var createdAt time.Time
	if resp.Data.CreatedAt != "" {
		createdAt, err = time.Parse(time.RFC3339, resp.Data.CreatedAt)
		if err != nil {
			createdAt, err = time.Parse("2006-01-02 15:04:05", resp.Data.CreatedAt)
			if err != nil {
				log.Errorf("[GRPCClient-5] GetProductByIDGRPC: failed to parse createdAt '%s': %v", resp.Data.CreatedAt, err)
				createdAt = time.Time{} // zero time
			}
		}
	}

	return &entity.ProductResponseEntity{
		ID:            productIDParsed, // Sudah aman
		ParentID:      parentIDParsed,  // Sudah aman (bisa uuid.Nil)
		ProductName:   resp.Data.Name,
		ProductImage:  resp.Data.Image,
		RegulerPrice:  float64(resp.Data.RegulerPrice),
		SalePrice:     float64(resp.Data.SalePrice),
		Unit:          resp.Data.Unit,
		Weight:        int(resp.Data.Weight),
		Stock:         int(resp.Data.Stock),
		ProductStatus: resp.Data.Status,
		CategoryName:  resp.Data.CategoryName,
		CreatedAt:     createdAt,
		Child:         childEntities,
	}, nil
}

func (g *GRPCClient) Close() {
	if g.userConn != nil {
		g.userConn.Close()
		log.Info("User service GRPC connection closed")
	}
	if g.productConn != nil {
		g.productConn.Close()
		log.Info("Product service GRPC connection closed")
	}
}

func NewGRPCClient(cfg *config.Config) (*GRPCClient, error) {

	userConn, err := grpc.Dial(cfg.App.UserServiceGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	productConn, err := grpc.Dial(cfg.App.ProductServiceGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		userConn.Close()
		return nil, err
	}

	return &GRPCClient{
		userClient:    userPB.NewUserServiceClient(userConn),
		productClient: productPB.NewProductServiceClient(productConn),
		cfg:           cfg,
	}, nil

}
