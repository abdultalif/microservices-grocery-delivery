package service

import (
	"context"

	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/core/domain/entity"
	"github.com/labstack/gommon/log"
)

func (o *OrderService) getInternalTokenGRPC() (string, error) {
	return o.grpcClient.GetInternalTokenGRPC()
}

func (o *OrderService) GetUserByIDGRPC(userID int64, accessToken string) (*entity.CustomerResponseEntity, error) {
	return o.grpcClient.GetUserByIDGRPC(userID, accessToken)
}

func (o *OrderService) getProductByIDGRPC(productID string, accessToken string) (*entity.ProductResponseEntity, error) {
	return o.grpcClient.GetProductByIDGRPC(productID, accessToken)
}

func (o *OrderService) GetAllgRPC(ctx context.Context, query entity.QueryStringEntity) ([]entity.OrderEntity, int64, int64, error) {
	result, count, total, err := o.elasticRepo.SearchOrderElastic(ctx, query)
	if err == nil {
		return result, count, total, nil
	} else {
		log.Errorf("[OrderService-1] GetAll: %v", err)
	}

	result, count, total, err = o.orderRepository.GetAll(ctx, query)
	if err != nil {
		log.Errorf("[OrderService-1] GetAll: %v", err)
		return nil, 0, 0, err
	}

	token, err := o.getInternalTokenGRPC()
	if err != nil {
		log.Errorf("[OrderService-1] GetAll: %v", err)
		return nil, 0, 0, err
	}

	for key, val := range result {
		userResponse, err := o.GetUserByIDGRPC(val.BuyerID, token)
		if err != nil {
			log.Errorf("[OrderService-2] GetAll: %v", err)
			return nil, 0, 0, err
		}

		result[key].BuyerName = userResponse.Name

		for key2, res := range val.OrderItems {
			productResponse, err := o.getProductByIDGRPC(res.ProductID.String(), token)
			if err != nil {
				log.Errorf("[OrderService-3] GetAll: %v", err)
				return nil, 0, 0, err
			}
			val.OrderItems[key2].ProductImage = productResponse.ProductImage
		}
	}
	return result, count, total, nil
}
