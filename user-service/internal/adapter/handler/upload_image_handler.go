package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/adapter/handler/response"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/adapter/storage"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/core/domain/entity"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/core/service"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type UploadImageHandlerInterface interface {
	UploadImage(c echo.Context) error
}

type uploadImageHandler struct {
	userService    service.UserServiceInterface
	storageHandler storage.SupabaseInterface
}

// UploadImage implements UploadImageHandlerInterface.
func (u *uploadImageHandler) UploadImage(c echo.Context) error {

	var (
		ctx = c.Request().Context()
		res = response.ResponseDefault{}
	)

	user, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[UserHandler-1] ChangePassword: user data not found in context")
		res.Success = false
		res.Code = http.StatusUnauthorized
		res.Message = "unauthorized"
		return c.JSON(http.StatusUnauthorized, res)
	}

	file, err := c.FormFile("photo")
	if err != nil {
		log.Errorf("[UploadImageHandler-1] UploadImage: %v", err)
		res.Message = err.Error()
		res.Success = false
		res.Code = http.StatusBadRequest
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	src, err := file.Open()
	if err != nil {
		log.Errorf("[UploadImageHandler-2] UploadImage: %v", err)
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		res.Success = false
		res.Data = nil
		return c.JSON(http.StatusBadRequest, res)
	}

	defer src.Close()

	fileBuffer := new(bytes.Buffer)
	_, err = io.Copy(fileBuffer, src)
	if err != nil {
		log.Errorf("[UploadImageHandler-3] UploadImage: %v", err)
		res.Code = http.StatusInternalServerError
		res.Message = err.Error()
		res.Success = false
		res.Data = nil
		return c.JSON(http.StatusInternalServerError, res)
	}

	newFileName := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), getExtention(file.Filename))

	uploadPath := fmt.Sprintf("public/uploads/%s", newFileName)

	url, err := u.storageHandler.UploadFile(uploadPath, fileBuffer)
	if err != nil {
		log.Errorf("[UploadImageHandler-4] UploadImage: %v", err)
		res.Code = http.StatusInternalServerError
		res.Message = err.Error()
		res.Success = false
		res.Data = nil
		return c.JSON(http.StatusInternalServerError, res)
	}

	err = u.userService.UploadPhoto(ctx, user.UserID, url)
	if err != nil {
		log.Errorf("[UploadImageHandler-5] UpdatePhoto: %v", err)
		res.Code = http.StatusInternalServerError
		res.Message = "Upload berhasil tapi gagal update photo ke database"
		res.Success = false
		res.Data = map[string]string{"image_url": url}
		return c.JSON(http.StatusInternalServerError, res)
	}

	res.Code = http.StatusOK
	res.Message = "Success"
	res.Success = true
	res.Data = map[string]string{"image_url": url}

	return c.JSON(http.StatusOK, res)
}

func getExtention(fileName string) string {
	ext := fileName[len(fileName)-3:]
	if len(fileName) > 4 && fileName[len(fileName)-4] == '.' {
		ext = fileName[len(fileName)-4:]
	}
	if ext[0] != '.' {
		ext = "." + ext
	}
	return ext
}

func NewUploadImageHandler(userService service.UserServiceInterface, storageHandler storage.SupabaseInterface) UploadImageHandlerInterface {
	return &uploadImageHandler{
		storageHandler: storageHandler,
		userService:    userService,
	}
}
