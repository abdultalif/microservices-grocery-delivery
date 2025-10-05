package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/adapter/handler/request"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/adapter/handler/response"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/core/domain/entity"
	errs "github.com/abdultalif/microservices-grocery-delivery/user-service/internal/core/domain/error"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/core/service"
	v "github.com/abdultalif/microservices-grocery-delivery/user-service/utils/validator"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type RoleHandlerInterface interface {
	CreateRole(c echo.Context) error
	GetAllRole(c echo.Context) error
	GetRoleByID(c echo.Context) error
	UpdateRole(c echo.Context) error
	DeleteRole(c echo.Context) error
}

type roleHandler struct {
	roleService service.RoleServiceInterface
}

// CreateRole implements RoleHandlerInterface.
func (r *roleHandler) CreateRole(c echo.Context) error {
	var (
		req = request.RoleRequest{}
		res = response.ResponseDefault{}
		ctx = c.Request().Context()
	)

	if err := c.Bind(&req); err != nil {
		log.Errorf("[RoleHandler-1] CreateRole: %v", err)
		res.Success = false
		res.Data = nil
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		return c.JSON(http.StatusBadRequest, res)
	}

	if err := c.Validate(req); err != nil {
		log.Errorf("[RoleHandler-2] CreateRole: %v", err)
		if ve, ok := err.(v.ValidationError); ok {
			res.Code = http.StatusUnprocessableEntity
			res.Success = false
			res.Data = nil
			res.Message = ve.Errors
			return c.JSON(http.StatusUnprocessableEntity, res)
		}
		res.Code = http.StatusUnprocessableEntity
		res.Success = false
		res.Data = nil
		res.Message = err.Error()
		return c.JSON(http.StatusUnprocessableEntity, res)
	}

	reqEntity := entity.RoleEntity{
		Name: req.Name,
	}

	if err := r.roleService.Create(ctx, reqEntity); err != nil {
		log.Errorf("[RoleHandler-3] CreateRole: %v", err)
		if errors.Is(err, errs.ErrUserExist) {
			res.Code = http.StatusConflict
			res.Success = false
			res.Data = nil
			res.Message = "Role already exists"
			return c.JSON(http.StatusConflict, res)
		}
		res.Code = http.StatusInternalServerError
		res.Success = false
		res.Data = nil
		res.Message = err.Error()
		return c.JSON(http.StatusInternalServerError, res)
	}

	res.Code = http.StatusCreated
	res.Success = true
	res.Message = "Role created successfully"
	return c.JSON(http.StatusCreated, res)
}

// DeleteRole implements RoleHandlerInterface.
func (r *roleHandler) DeleteRole(c echo.Context) error {

	var (
		res = response.ResponseDefault{}
		ctx = c.Request().Context()
	)

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		log.Errorf("[RoleHandler-1] DeleteRole: %v", err)
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	if err = r.roleService.Delete(ctx, userID); err != nil {
		log.Errorf("[RoleHandler-2] DeleteRole: %v", err)
		if errors.Is(err, errs.ErrRoleHasUsers) {
			res.Message = "Role has users"
			res.Success = false
			res.Code = http.StatusBadRequest
			res.Data = nil
			return c.JSON(http.StatusBadRequest, res)
		} else if errors.Is(err, errs.ErrUserNotFound) {
			res.Message = "Role not found"
			res.Success = false
			res.Code = http.StatusNotFound
			res.Data = nil
			return c.JSON(http.StatusNotFound, res)
		} else {
			res.Message = err.Error()
			res.Success = false
			res.Code = http.StatusInternalServerError
			res.Data = nil
			return c.JSON(http.StatusInternalServerError, res)
		}
	}

	res.Message = "Role deleted successfully"
	res.Success = true
	res.Code = http.StatusOK
	res.Data = nil
	return c.JSON(http.StatusOK, res)

}

// GetAllRole implements RoleHandlerInterface.
func (r *roleHandler) GetAllRole(c echo.Context) error {
	var (
		res     = response.ResponseDefault{}
		ctx     = c.Request().Context()
		resRole = []response.RoleResponse{}
	)

	search := c.QueryParam("search")

	role, err := r.roleService.GetAll(ctx, search)
	if err != nil {
		log.Errorf("[RoleHandler-1] GetAllRole: %v", err)
		res.Success = false
		res.Code = http.StatusInternalServerError
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusInternalServerError, res)
	}

	for _, v := range role {
		resRole = append(resRole, response.RoleResponse{
			ID:   v.ID,
			Name: v.Name,
		})
	}

	res.Message = "Role found successfully"
	res.Success = true
	res.Code = http.StatusOK
	res.Data = resRole
	return c.JSON(http.StatusOK, res)

}

// GetRoleByID implements RoleHandlerInterface.
func (r *roleHandler) GetRoleByID(c echo.Context) error {

	var (
		res     = response.ResponseDefault{}
		ctx     = c.Request().Context()
		resRole = response.RoleResponse{}
	)

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		log.Errorf("[RoleHandler-3] GetRoleByID: %v", err)
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	role, err := r.roleService.GetByID(ctx, userID)
	if err != nil {
		log.Errorf("[RoleHandler-4] GetRoleByID: %v", err)
		if errors.Is(err, errs.ErrUserNotFound) {
			res.Message = "Role not found"
			res.Success = false
			res.Code = http.StatusNotFound
			res.Data = nil
			return c.JSON(http.StatusNotFound, res)
		}
		res.Success = false
		res.Code = http.StatusInternalServerError
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusInternalServerError, res)
	}

	resRole.ID = role.ID
	resRole.Name = role.Name

	res.Message = "Role found successfully"
	res.Success = true
	res.Code = http.StatusOK
	res.Data = resRole
	return c.JSON(http.StatusOK, res)

}

// UpdateRole implements RoleHandlerInterface.
func (r *roleHandler) UpdateRole(c echo.Context) error {
	var (
		req = request.RoleRequest{}
		res = response.ResponseDefault{}
		ctx = c.Request().Context()
	)

	if err := c.Bind(&req); err != nil {
		log.Errorf("[RoleHandler-3] UpdateRole: %v", err)
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		log.Errorf("[RoleHandler-4] UpdateRole: %v", err)
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	if err = c.Validate(req); err != nil {
		log.Errorf("[RoleHandler-5] UpdateRole: %v", err)

		if ve, ok := err.(v.ValidationError); ok {
			res.Message = ve.Errors
			res.Success = false
			res.Code = http.StatusUnprocessableEntity
			res.Data = nil
			return c.JSON(http.StatusUnprocessableEntity, res)
		}

		res.Message = err.Error()
		res.Success = false
		res.Code = http.StatusUnprocessableEntity
		res.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, res)
	}

	reqEntity := entity.RoleEntity{
		ID:   userID,
		Name: req.Name,
	}

	err = r.roleService.Update(ctx, reqEntity)
	if err != nil {
		log.Errorf("[RoleHandler-6] UpdateRole: %v", err)
		if err.Error() == "404" {
			res.Message = "Role not found"
			res.Success = false
			res.Code = http.StatusNotFound
			res.Data = nil
			return c.JSON(http.StatusNotFound, res)
		}
		res.Success = false
		res.Code = http.StatusInternalServerError
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusInternalServerError, res)
	}

	res.Message = "Role updated successfully"
	res.Success = true
	res.Code = http.StatusOK
	res.Data = nil
	return c.JSON(http.StatusOK, res)

}

func NewRoleHandler(roleService service.RoleServiceInterface) RoleHandlerInterface {
	return &roleHandler{roleService: roleService}
}
