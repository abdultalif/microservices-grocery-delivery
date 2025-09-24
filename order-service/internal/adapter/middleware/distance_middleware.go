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

type middlewareDistanceInterface interface {
	DistanceCheck() echo.MiddlewareFunc
}

type middlewareDistance struct {
	cfg *config.Config
}

// haversineDistance menghitung jarak antara dua koordinat (lat, lng) dalam kilometer
// menggunakan rumus Haversine.
// NOTE: Perbaikan dari kode sebelumnya:
// 1. Konversi derajat ke radian pakai (deg * math.Pi / 180), bukan /100.
// 2. Penempatan kurung di math.Cos sudah diperbaiki agar sesuai rumus.
func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371 // jari-jari bumi dalam kilometer

	// konversi selisih koordinat ke radian
	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLng := (lng2 - lng1) * (math.Pi / 180)

	// konversi koordinat awal & tujuan ke radian
	lat1Rad := lat1 * (math.Pi / 180)
	lat2Rad := lat2 * (math.Pi / 180)

	// rumus haversine
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// DistanceCheck middleware untuk validasi apakah user berada dalam radius MAX_DISTANCE
func (m *middlewareDistance) DistanceCheck() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Ambil query parameter
			latParam := c.QueryParam("lat")
			lonParam := c.QueryParam("lng")

			if latParam == "" || lonParam == "" {
				log.Errorf("[MiddlewareDistance] Missing or invalid lat/lng")
				return c.JSON(http.StatusBadRequest, response.ResponseAPI(
					false,
					http.StatusBadRequest,
					"Missing or invalid lat or lng",
					nil,
				))
			}

			// Konversi ke float64
			lat, err1 := strconv.ParseFloat(latParam, 64)
			lng, err2 := strconv.ParseFloat(lonParam, 64)

			if err1 != nil || err2 != nil {
				log.Errorf("[MiddlewareDistance] Failed parse lat/lng")
				return c.JSON(http.StatusBadRequest, response.ResponseAPI(
					false,
					http.StatusBadRequest,
					"Missing or invalid lat or lng",
					nil,
				))
			}

			// Ambil referensi dari config (.env)
			latRef, _ := strconv.ParseFloat(m.cfg.App.LatitudeRef, 64)
			lngRef, _ := strconv.ParseFloat(m.cfg.App.LongitudeRef, 64)

			// Hitung jarak
			distance := haversineDistance(latRef, lngRef, lat, lng)

			// Bandingkan dengan batas maksimal
			if distance > float64(m.cfg.App.MaxDistance) {
				log.Errorf("[MiddlewareDistance] Distance too far: %.2f km", distance)
				return c.JSON(http.StatusBadRequest, response.ResponseAPI(
					false,
					http.StatusBadRequest,
					"Distance too far",
					nil,
				))
			}

			// Kalau jarak masih dalam batas, lanjut ke handler berikutnya
			return next(c)
		}
	}
}

func NewMiddlewareDistance(cfg *config.Config) middlewareDistanceInterface {
	return &middlewareDistance{
		cfg: cfg,
	}
}
