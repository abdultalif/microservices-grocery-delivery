package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/abdultalif/microservices-grocery-delivery/order-service/config"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type HttpClient interface {
	Connect()
	CallURL(method, url string, header map[string]string, rawData []byte) (*http.Response, error)
}

type Options struct {
	timeout int
	http    *http.Client
	logger  echo.Logger
}

type loggingTransport struct {
	logger echo.Logger
}

// CallURL implements HttpClient.
func (o *Options) CallURL(method, url string, header map[string]string, rawData []byte) (*http.Response, error) {

	o.Connect()

	req, err := http.NewRequest(method, url, bytes.NewBuffer(rawData))
	if err != nil {
		return nil, fmt.Errorf("httpclient: build request failed: %w", err)
	}

	for key, value := range header {
		req.Header.Set(key, value)
	}

	res, err := o.http.Do(req)
	if err != nil {

		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, fmt.Errorf("timeout: %w", err)
		}

		if strings.Contains(err.Error(), "connection refused") {
			return nil, fmt.Errorf("connection refused: %w", err)
		}

		return nil, fmt.Errorf("http call failed: %w", err)
	}

	if res.StatusCode >= 500 {
		return nil, fmt.Errorf("server error: %d", res.StatusCode)
	}

	if res.StatusCode == 404 {
		return res, fmt.Errorf("not found")
	}

	if res.StatusCode >= 400 {
		return res, fmt.Errorf("client error: %d", res.StatusCode)
	}

	return res, nil
}

// Connect implements HttpClient.
func (o *Options) Connect() {

	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	httpClient := &http.Client{
		Timeout:   time.Duration(o.timeout) * time.Second,
		Transport: &loggingTransport{e.Logger},
	}

	o.http = httpClient
	o.logger = e.Logger
}

func NewHttpClient(cfg *config.Config) HttpClient {
	opt := new(Options)
	opt.timeout = cfg.App.ServerTimeOut
	return opt
}

func (lt *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {

	lt.logger.Infof("Making request to: %s %s", req.Method, req.URL)
	lt.logger.Infof("Request Headers: %+v", req.Header)

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	lt.logger.Infof("Request Body: %s", req.Body)

	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		lt.logger.Infof("Request Failed: %v", err)
		return nil, err
	}

	lt.logger.Infof("Received response with status: %s", res.Status)
	lt.logger.Infof("Response Headers: %+v", res.Header)

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		lt.logger.Infof("Response Body: %+v", resBody)
	}
	res.Body = io.NopCloser(bytes.NewBuffer(resBody))
	return res, nil
}
