package middleware

import (
	"math"
	"net/http"
	"order-service/config"
	"order-service/internal/adapter/handler/response"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type middlewareDistanceIterface interface {
	DistanceCheck() echo.MiddlewareFunc
}

type middlewareDistance struct {
	cfg *config.Config
}

func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	// Radius bumi dalam kilometer
	const R = 6371 

	dLat := (lat2 - lat1) * (math.Pi / 100)
	dLng := (lng2 - lng1) * (math.Pi / 100)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) + 
		math.Cos(lat1*(math.Pi/180)*math.Cos(lat2*(math.Pi/180))*
			math.Sin(dLng/2)*math.Sin(dLng/2))
	
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// DistanceCheck implements middlewareDistanceIterface.
func (m *middlewareDistance) DistanceCheck() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			latParam := c.QueryParam("lat")
			lonParam := c.QueryParam("lng")
			if latParam == "" || lonParam == "" {
				log.Errorf("[MiddlewareDistance-1] DistanceCheck: %s", "Missing or invalid lat or lng")
				return c.JSON(http.StatusBadRequest, response.ResponseAPI(false, http.StatusBadRequest, "Missing or invalid lat or lng", nil))				
			}
			
			lat, err1 := strconv.ParseFloat(latParam, 64)
			lng, err2 := strconv.ParseFloat(lonParam, 64)
			
			if err1 != nil || err2 != nil {
				log.Errorf("[MiddlewareDistance-1] DistanceCheck: %s", "Missing or invalid lat or lng")
				return c.JSON(http.StatusBadRequest, response.ResponseAPI(false, http.StatusBadRequest, "Missing or invalid lat or lng", nil))
			}

			latRef, _ := strconv.ParseFloat(m.cfg.App.LatitudeRef, 64)
			lngRef, _ := strconv.ParseFloat(m.cfg.App.LongitudeRef, 64)
			distance := haversineDistance(latRef, lngRef, lat, lng)
			if distance > float64(m.cfg.App.MaxDistance) {
				log.Errorf("[MiddlewareDistance-1] DistanceCheck: %s", "Distance too far")
				return c.JSON(http.StatusBadRequest, response.ResponseAPI(false, http.StatusBadRequest, "Distance too far", nil))
			}

			return next(c)

		}
	}
}

func NewmiddlewareDistance(cfg *config.Config) middlewareDistanceIterface {
	return &middlewareDistance{
		cfg: cfg,
	}
}
