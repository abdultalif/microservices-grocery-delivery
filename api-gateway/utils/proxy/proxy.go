package proxy

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/labstack/echo/v4"
)

func ForwardRequest(c echo.Context, targetURL string) error {
	target, err := url.Parse(targetURL)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "Invalid target URL")
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to read request body")
	}

	req, err := http.NewRequest(c.Request().Method, target.String(), bytes.NewReader(body))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create request")
	}

	for key, values := range c.Request().Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	req.Header.Set("X-API-Gateway", "true")
	req.Header.Set("X-API-Gateway-Version", "1.0")
	req.Header.Set("X-Request-ID", c.Response().Header().Get("X-Request-ID"))

	req.Header.Set("X-Forwarded-For", c.RealIP())
	req.Header.Set("X-Real-IP", c.RealIP())

	gatewaySecret := os.Getenv("GATEWAY_SECRET_KEY")
	if gatewaySecret != "" {
		req.Header.Set("X-Gateway-Secret", gatewaySecret)
	}

	req.URL.RawQuery = c.Request().URL.RawQuery

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to forward request")
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to read response body")
	}

	for key, values := range resp.Header {
		for _, value := range values {
			c.Response().Header().Add(key, value)
		}
	}

	return c.Blob(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)

}
