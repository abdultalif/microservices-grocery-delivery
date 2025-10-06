package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errs "github.com/abdultalif/microservices-grocery-delivery/user-service/internal/core/domain/error"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/core/service"
	userPB "github.com/abdultalif/microservices-grocery-delivery/user-service/proto/user"
	"github.com/labstack/gommon/log"
)

type GRPCUserHandler struct {
	userPB.UnimplementedUserServiceServer
	customerService service.CustomerServiceInterface
	jwtService      service.JwtServiceInterface
}

func NewGRPCUserHandler(customerService service.CustomerServiceInterface, jwtService service.JwtServiceInterface) *GRPCUserHandler {
	return &GRPCUserHandler{
		customerService: customerService,
		jwtService:      jwtService,
	}
}

func (g *GRPCUserHandler) GetCustomerByID(ctx context.Context, req *userPB.GetCustomerRequest) (*userPB.GetCustomerResponse, error) {

	result, err := g.customerService.GetCustomerByID(ctx, req.UserId)
	if err != nil {
		log.Errorf("[GRPCUserHandler-1] GetCustomerByID: %v", err)
		if errors.Is(err, errs.ErrUserNotFound) {
			return &userPB.GetCustomerResponse{
				Success: false,
				Code:    404,
				Message: "Customer not found",
				Data:    nil,
			}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userPB.GetCustomerResponse{
		Success: true,
		Code:    200,
		Message: "success",
		Data: &userPB.CustomerData{
			Id:       result.ID,
			Name:     result.Name,
			Email:    result.Email,
			Phone:    result.Phone,
			Address:  result.Address,
			Photo:    result.Photo,
			Lat:      result.Lat,
			Lng:      result.Lng,
			RoleId:   result.RoleID,
			RoleName: result.RoleName,
		},
	}, nil
}

func (g *GRPCUserHandler) GetInternalToken(ctx context.Context, req *userPB.GetInternalTokenRequest) (*userPB.GetInternalTokenResponse, error) {
	token, err := g.jwtService.GenerateToken(0)
	if err != nil {
		log.Errorf("[GRPCUserHandler-1] GetInternalToken: %v", err)
		return &userPB.GetInternalTokenResponse{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		}, nil
	}

	return &userPB.GetInternalTokenResponse{
		Success: true,
		Message: "success",
		Data: &userPB.TokenData{
			AccessToken: token,
		},
	}, nil
}
