package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"user-service/config"
	"user-service/internal/adapter"
	"user-service/internal/adapter/handler/request"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/service"
	v "user-service/utils/validator"

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
		jwtUserData = entity.JwtUserData{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[RoleHandler-1] CreateRole: %s", "data token not found")
		res.Success = false
		res.Code = http.StatusNotFound
		res.Message = "data token not found"
		res.Data = nil
		return c.JSON(http.StatusNotFound, res)
	}

	err := json.Unmarshal([]byte(user), &jwtUserData)
	if err != nil {
		log.Errorf("[RoleHandler-2] CreateRole: %v", err)
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	if err = c.Bind(&req); err != nil {
		log.Errorf("[RoleHandler-3] CreateRole: %v", err)
		res.Message = "Invalid request body format"
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	if err = c.Validate(req); err != nil {
		log.Errorf("[UserHandler-4] UpdateDataUser: %v", err)

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
		Name: req.Name,
	}

	if err = r.roleService.Create(ctx, reqEntity); err != nil {
		if err.Error() == "400" {
			res.Message = "Role already exists"
			res.Success = false
			res.Code = http.StatusBadRequest
			res.Data = nil
			return c.JSON(http.StatusBadRequest, res)
		}
		log.Errorf("[RoleHandler-5] CreateRole: %v", err)
		res.Message = err.Error()
		res.Success = false
		res.Code = http.StatusInternalServerError
		res.Data = nil
		return c.JSON(http.StatusInternalServerError, res)
	}

	res.Message = "Role created successfully"
	res.Success = true
	res.Code = http.StatusCreated
	res.Data = nil
	return c.JSON(http.StatusCreated, res)
}

// DeleteRole implements RoleHandlerInterface.
func (r *roleHandler) DeleteRole(c echo.Context) error {
	
	var (
		res = response.ResponseDefault{}
		ctx = c.Request().Context()
		jwtUserData = entity.JwtUserData{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[RoleHandler-1] DeleteRole: %s", "data token not found")
		res.Success = false
		res.Code = http.StatusNotFound
		res.Message = "data token not found"
		res.Data = nil
		return c.JSON(http.StatusNotFound, res)
	}

	err := json.Unmarshal([]byte(user), &jwtUserData)
	if err != nil {
		log.Errorf("[RoleHandler-2] DeleteRole: %v", err)
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		log.Errorf("[RoleHandler-3] DeleteRole: %v", err)
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	if err = r.roleService.Delete(ctx, userID); err != nil {
		log.Errorf("[RoleHandler-3] DeleteRole: %v", err)
		if err.Error() == "400" {
			res.Message = "Role has users"
			res.Success = false
			res.Code = http.StatusBadRequest
			res.Data = nil
			return c.JSON(http.StatusBadRequest, res)
		} else if err.Error() == "404" {
			res.Message = "Role not found"
			res.Success = false
			res.Code = http.StatusNotFound
			res.Data = nil
			return c.JSON(http.StatusNotFound, res)
		} else  {
			log.Errorf("[RoleHandler-3] DeleteRole: %v", err)
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
		res = response.ResponseDefault{}
		ctx = c.Request().Context()
		// jwtUserData = entity.JwtUserData{}
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
		res = response.ResponseDefault{}
		ctx = c.Request().Context()
		jwtUserData = entity.JwtUserData{}
		resRole = response.RoleResponse{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[RoleHandler-1] GetRoleByID: %s", "data token not found")
		res.Success = false
		res.Code = http.StatusNotFound
		res.Message = "data token not found"
		res.Data = nil
		return c.JSON(http.StatusNotFound, res)
	}

	err := json.Unmarshal([]byte(user), &jwtUserData)
	if err != nil {
		log.Errorf("[RoleHandler-2] GetRoleByID: %v", err)
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

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
		jwtUserData = entity.JwtUserData{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[RoleHandler-1] UpdateRole: %s", "data token not found")
		res.Success = false
		res.Code = http.StatusNotFound
		res.Message = "data token not found"
		res.Data = nil
		return c.JSON(http.StatusNotFound, res)
	}

	err := json.Unmarshal([]byte(user), &jwtUserData)
	if err != nil {
		log.Errorf("[RoleHandler-2] UpdateRole: %v", err)
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

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

func NewRoleHandler(roleService service.RoleServiceInterface, g *echo.Group, cfg *config.Config, jwtService service.JwtServiceInterface) RoleHandlerInterface {
	roleHandler := &roleHandler{roleService: roleService}

	mid := adapter.NewMiddlewareAdapter(cfg, jwtService)
	g.Use(mid.CheckToken())
	adminGroup := g.Group("/admin", mid.CheckToken())

	adminGroup.GET("/role", roleHandler.GetAllRole)
	adminGroup.POST("/role", roleHandler.CreateRole)
	adminGroup.DELETE("/role/:id", roleHandler.DeleteRole)
	adminGroup.GET("/role/:id", roleHandler.GetRoleByID)
	adminGroup.PATCH("/role/:id", roleHandler.UpdateRole)

	return roleHandler
}
