package middleware

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/abdultalif/microservices-grocery-delivery/order-service/config"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/handler/response"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/message"
	coreRepo "github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/repository"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/core/domain/entity"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type MiddlewareDistanceInterface interface {
	DistanceCheck() echo.MiddlewareFunc
}

type middlewareDistance struct {
	cfg           *config.Config
	localDataRepo coreRepo.LocalDataRepositoryInterface
	publisher     message.PublishRabbitMQInterface
}

func NewMiddlewareDistance(
	cfg *config.Config,
	localDataRepo coreRepo.LocalDataRepositoryInterface,
	publisher message.PublishRabbitMQInterface,
) MiddlewareDistanceInterface {
	return &middlewareDistance{
		cfg:           cfg,
		localDataRepo: localDataRepo,
		publisher:     publisher,
	}
}

func (m *middlewareDistance) DistanceCheck() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			userID := getUserIDFromContext(c)
			if userID == 0 {
				log.Errorf("[MiddlewareDistance] User ID not found in context")
				return c.JSON(http.StatusUnauthorized, response.ResponseAPI(
					false, http.StatusUnauthorized, "User not authenticated", nil,
				))
			}

			latParam := c.QueryParam("lat")
			lngParam := c.QueryParam("lng")

			var userLat, userLng float64
			isFirstTimeLocation := false

			if latParam != "" && lngParam != "" {
				lat, err1 := strconv.ParseFloat(latParam, 64)
				lng, err2 := strconv.ParseFloat(lngParam, 64)
				if err1 != nil || err2 != nil {
					return c.JSON(http.StatusBadRequest, response.ResponseAPI(
						false, http.StatusBadRequest, "Invalid lat or lng format", nil,
					))
				}

				// Update lokasi ke local DB langsung — tanpa HTTP
				if err := m.localDataRepo.UpdateBuyerLocation(ctx, userID, latParam, lngParam); err != nil {
					log.Warnf("[MiddlewareDistance] Gagal update lokasi ke local DB: %v", err)
				} else {
					log.Infof("[MiddlewareDistance] Lokasi user %d diupdate ke local DB", userID)
				}

				// Publish async ke user service via RabbitMQ — fire and forget
				go func() {
					if err := m.publisher.PublishUpdateUserLocation(userID, latParam, lngParam); err != nil {
						log.Warnf("[MiddlewareDistance] Gagal publish update lokasi ke RabbitMQ: %v", err)
					}
				}()

				userLat, userLng = lat, lng
				isFirstTimeLocation = true

			} else {
				// Ambil lokasi dari local DB — tanpa HTTP
				buyer, err := m.localDataRepo.GetBuyer(ctx, userID)
				if err != nil {
					log.Errorf("[MiddlewareDistance] Buyer %d tidak ada di local DB: %v", userID, err)
					return c.JSON(http.StatusBadRequest, response.ResponseAPI(
						false,
						http.StatusBadRequest,
						"Please set your location first (lat and lng required)",
						nil,
					))
				}

				if buyer.Lat == "" || buyer.Lng == "" {
					return c.JSON(http.StatusBadRequest, response.ResponseAPI(
						false,
						http.StatusBadRequest,
						"User location not set, please provide lat and lng",
						nil,
					))
				}

				lat, err1 := strconv.ParseFloat(buyer.Lat, 64)
				lng, err2 := strconv.ParseFloat(buyer.Lng, 64)
				if err1 != nil || err2 != nil {
					return c.JSON(http.StatusInternalServerError, response.ResponseAPI(
						false, http.StatusInternalServerError, "Invalid lat/lng format in database", nil,
					))
				}

				userLat, userLng = lat, lng
				log.Infof("[MiddlewareDistance] Lokasi user %d diambil dari local DB", userID)
			}

			// Hitung jarak
			latRef, err := strconv.ParseFloat(m.cfg.App.LatitudeRef, 64)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, response.ResponseAPI(
					false, http.StatusInternalServerError, "Service configuration error", nil,
				))
			}

			lngRef, err := strconv.ParseFloat(m.cfg.App.LongitudeRef, 64)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, response.ResponseAPI(
					false, http.StatusInternalServerError, "Service configuration error", nil,
				))
			}

			distance := haversineDistance(latRef, lngRef, userLat, userLng)

			if distance > float64(m.cfg.App.MaxDistance) {
				log.Errorf("[MiddlewareDistance] Jarak terlalu jauh: %.2f km (max: %d km)",
					distance, m.cfg.App.MaxDistance)
				return c.JSON(http.StatusBadRequest, response.ResponseAPI(
					false,
					http.StatusBadRequest,
					"Your location is outside our service area",
					map[string]interface{}{
						"distance":     fmt.Sprintf("%.2f km", distance),
						"max_distance": fmt.Sprintf("%d km", m.cfg.App.MaxDistance),
					},
				))
			}

			c.Set("user_distance", distance)
			c.Set("user_lat", userLat)
			c.Set("user_lng", userLng)
			c.Set("is_first_time_location", isFirstTimeLocation)

			log.Infof("[MiddlewareDistance] User %d jarak %.2f km dari service center (max: %d km)",
				userID, distance, m.cfg.App.MaxDistance)

			return next(c)
		}
	}
}

func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371
	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLng := (lng2 - lng1) * (math.Pi / 180)
	lat1Rad := lat1 * (math.Pi / 180)
	lat2Rad := lat2 * (math.Pi / 180)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func getUserIDFromContext(c echo.Context) int64 {
	userData, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Errorf("[MiddlewareDistance] Failed to get user data from context")
		return 0
	}
	return userData.UserID
}

// package middleware

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"math"
// 	"net/http"
// 	"strconv"

// 	"github.com/abdultalif/microservices-grocery-delivery/order-service/config"
// 	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/handler/response"
// 	httpclient "github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/http_client"
// 	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/core/domain/entity"
// 	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/core/service"

// 	"github.com/labstack/echo/v4"
// 	"github.com/labstack/gommon/log"
// )

// type MiddlewareDistanceInterface interface {
// 	DistanceCheck() echo.MiddlewareFunc
// 	getUserLocation(userID int64, token string) (lat, lng float64, err error)
// 	updateUserLocation(userID int64, token, lat, lng string) error
// }

// type middlewareDistance struct {
// 	cfg          *config.Config
// 	httpClient   httpclient.HttpClient
// 	orderService service.OrderServiceInterface
// }

// // DistanceCheck implements middlewareDistanceInterface.
// func (m *middlewareDistance) DistanceCheck() echo.MiddlewareFunc {
// 	return func(next echo.HandlerFunc) echo.HandlerFunc {
// 		return func(c echo.Context) error {

// 			userID := getUserIDFromContext(c)
// 			if userID == 0 {
// 				log.Errorf("[MiddlewareDistance] User ID not found in context")
// 				return c.JSON(http.StatusUnauthorized, response.ResponseAPI(
// 					false,
// 					http.StatusUnauthorized,
// 					"User not authenticated",
// 					nil,
// 				))
// 			}

// 			token, err := m.orderService.GetInternalToken()
// 			if err != nil {
// 				log.Errorf("[OrderService-1] CreateOrder: %v", err)
// 				return c.JSON(http.StatusInternalServerError, response.ResponseAPI(false, http.StatusInternalServerError, err.Error(), nil))
// 			}

// 			latParam := c.QueryParam("lat")
// 			lonParam := c.QueryParam("lng")

// 			var userLat, userLng float64
// 			isFirstTimeLocation := false

// 			if latParam != "" && lonParam != "" {
// 				lat, err1 := strconv.ParseFloat(latParam, 64)
// 				lng, err2 := strconv.ParseFloat(lonParam, 64)
// 				if err1 != nil || err2 != nil {
// 					return c.JSON(http.StatusBadRequest, response.ResponseAPI(
// 						false,
// 						http.StatusBadRequest,
// 						"Invalid lat or lng format",
// 						nil,
// 					))
// 				}

// 				if err := m.updateUserLocation(userID, token, latParam, lonParam); err != nil {
// 					log.Errorf("[MiddlewareDistance] Failed to update user location: %v", err)
// 				} else {
// 					log.Infof("[MiddlewareDistance] Successfully updated location for user %d", userID)
// 				}

// 				userLat, userLng = lat, lng
// 				isFirstTimeLocation = true

// 			} else {
// 				userLat, userLng, err = m.getUserLocation(userID, token)
// 				if err != nil {
// 					log.Errorf("[MiddlewareDistance] Failed to get user location: %v", err)
// 					return c.JSON(http.StatusBadRequest, response.ResponseAPI(
// 						false,
// 						http.StatusBadRequest,
// 						"Please set your location first (lat and lng required)",
// 						nil,
// 					))
// 				}
// 				log.Infof("[MiddlewareDistance] Using saved location for user %d", userID)
// 			}

// 			latRef, err := strconv.ParseFloat(m.cfg.App.LatitudeRef, 64)
// 			if err != nil {
// 				log.Errorf("[MiddlewareDistance] Invalid LatitudeRef in config: %v", err)
// 				return c.JSON(http.StatusInternalServerError, response.ResponseAPI(
// 					false,
// 					http.StatusInternalServerError,
// 					"Service configuration error",
// 					nil,
// 				))
// 			}

// 			lngRef, err := strconv.ParseFloat(m.cfg.App.LongitudeRef, 64)
// 			if err != nil {
// 				log.Errorf("[MiddlewareDistance] Invalid LongitudeRef in config: %v", err)
// 				return c.JSON(http.StatusInternalServerError, response.ResponseAPI(
// 					false,
// 					http.StatusInternalServerError,
// 					"Service configuration error",
// 					nil,
// 				))
// 			}

// 			distance := haversineDistance(latRef, lngRef, userLat, userLng)

// 			if distance > float64(m.cfg.App.MaxDistance) {
// 				log.Errorf("[MiddlewareDistance] Distance too far: %.2f km (max: %d km)",
// 					distance, m.cfg.App.MaxDistance)
// 				return c.JSON(http.StatusBadRequest, response.ResponseAPI(
// 					false,
// 					http.StatusBadRequest,
// 					"Your location is outside our service area",
// 					map[string]interface{}{
// 						"distance":     fmt.Sprintf("%.2f km", distance),
// 						"max_distance": fmt.Sprintf("%d km", m.cfg.App.MaxDistance),
// 					},
// 				))
// 			}

// 			c.Set("user_distance", distance)
// 			c.Set("user_lat", userLat)
// 			c.Set("user_lng", userLng)
// 			c.Set("is_first_time_location", isFirstTimeLocation)

// 			log.Infof("[MiddlewareDistance] User %d is %.2f km away from service center (max: %d km)",
// 				userID, distance, m.cfg.App.MaxDistance)

// 			return next(c)
// 		}
// 	}
// }

// // updateUserLocation implements middlewareDistanceInterface.
// func (m *middlewareDistance) updateUserLocation(userID int64, token string, lat string, lng string) error {
// 	// url := fmt.Sprintf("%s/admin/customers/%d/location", m.cfg.App.UserServiceUrl, userID) menggunakan port user-service
// 	url := fmt.Sprintf("%s/admin/customers/%d/location", m.cfg.App.ApiGatewayServiceUrl, userID) // menggunakan port apigateway-service

// 	reqBody := entity.UpdateLocationRequest{
// 		Lat: lat,
// 		Lng: lng,
// 	}

// 	headers := map[string]string{
// 		"Authorization": "Bearer " + token,
// 		"Content-Type":  "application/json",
// 		"Accept":        "application/json",
// 	}

// 	reqBodyBytes, err := json.Marshal(reqBody)
// 	if err != nil {
// 		log.Errorf("[MiddlewareDistance] Failed to marshal request body: %v", err)
// 		return err
// 	}

// 	resp, err := m.httpClient.CallURL("PUT", url, headers, reqBodyBytes)
// 	if err != nil {
// 		log.Errorf("[MiddlewareDistance] Failed to update user location: %v", err)
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		body, _ := io.ReadAll(resp.Body)
// 		log.Errorf("[MiddlewareDistance] Update location failed with status %d: %s",
// 			resp.StatusCode, string(body))
// 		return fmt.Errorf("failed to update user location")
// 	}

// 	return nil
// }

// func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
// 	const R = 6371 // jari-jari bumi dalam kilometer

// 	// konversi selisih koordinat ke radian
// 	dLat := (lat2 - lat1) * (math.Pi / 180)
// 	dLng := (lng2 - lng1) * (math.Pi / 180)

// 	// konversi koordinat awal & tujuan ke radian
// 	lat1Rad := lat1 * (math.Pi / 180)
// 	lat2Rad := lat2 * (math.Pi / 180)

// 	// rumus haversine
// 	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
// 		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
// 			math.Sin(dLng/2)*math.Sin(dLng/2)

// 	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

// 	return R * c
// }

// func (m *middlewareDistance) getUserLocation(userID int64, token string) (lat, lng float64, err error) {
// 	// url := fmt.Sprintf("%s/admin/customers/%d", m.cfg.App.UserServiceUrl, userID)
// 	url := fmt.Sprintf("%s/admin/customers/%d", m.cfg.App.ApiGatewayServiceUrl, userID) // menggunakan port apigateway-service

// 	headers := map[string]string{
// 		"Authorization": "Bearer " + token,
// 		"Accept":        "application/json",
// 	}

// 	resp, err := m.httpClient.CallURL("GET", url, headers, nil)
// 	if err != nil {
// 		log.Errorf("[MiddlewareDistance] HTTP call failed: %v", err)
// 		return 0, 0, err
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Errorf("[MiddlewareDistance] Failed to read response: %v", err)
// 		return 0, 0, err
// 	}

// 	var userResp entity.UserHttpClientResponse
// 	if err := json.Unmarshal(body, &userResp); err != nil {
// 		log.Errorf("[MiddlewareDistance] Failed to unmarshal response: %v", err)
// 		return 0, 0, err
// 	}

// 	if !userResp.Success {
// 		return 0, 0, fmt.Errorf("user service error: %s", userResp.Message)
// 	}

// 	if userResp.Data.Lat == "" || userResp.Data.Lng == "" {
// 		return 0, 0, fmt.Errorf("user location not set")
// 	}

// 	lat, err1 := strconv.ParseFloat(userResp.Data.Lat, 64)
// 	lng, err2 := strconv.ParseFloat(userResp.Data.Lng, 64)

// 	if err1 != nil || err2 != nil {
// 		return 0, 0, fmt.Errorf("invalid lat/lng format in database")
// 	}

// 	return lat, lng, nil
// }

// func NewMiddlewareDistance(cfg *config.Config, orderService service.OrderServiceInterface, httpClient httpclient.HttpClient) MiddlewareDistanceInterface {
// 	return &middlewareDistance{
// 		cfg:          cfg,
// 		orderService: orderService,
// 		httpClient:   httpClient,
// 	}
// }

// func getUserIDFromContext(c echo.Context) int64 {
// 	userData, ok := c.Get("user").(entity.JwtUserData)
// 	if !ok {
// 		log.Errorf("[MiddlewareDistance] Failed to get user data from context")
// 		return 0
// 	}
// 	return userData.UserID
// }
