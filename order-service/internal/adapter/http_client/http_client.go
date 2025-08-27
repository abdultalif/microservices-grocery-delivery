package httpclient

import (
	"bytes"
	"io"
	"net/http"
	"order-service/config"
	"time"

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
	logger  echo.Logger
}



// CallURL implements HttpClient.
func (o *Options) CallURL(method string, url string, header map[string]string, rawData []byte) (*http.Response, error) {
	
	o.Connect()
	req, err := http.NewRequest(method, url, bytes.NewBuffer(rawData))
	if err != nil {
		o.logger.Errorj(log.JSON{
			"message": "[CallURL-1] Failed To Prepare Request Client HTTP",
			"error": err.Error(),
		})
		return nil, err
	}

	if len(header) > 0 {
		for key, value := range header {
			req.Header.Set(key, value)
		}
	}

	res, err := o.http.Do(req)
	if err != nil {
		o.logger.Errorj(log.JSON{
			"message": "[CallURL-2] Failed To Do Request Client HTTP",
			"error": err.Error(),
		})
		return nil, err
	}

	return res, nil
}

// Connect implements HttpClient.
func (o *Options) Connect() {
	
	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	httpClient := &http.Client{
		Timeout: time.Duration(o.timeout) * time.Second,
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

	// logging sebelum request 
	lt.logger.Infof("Making request to: %s %s", req.Method, req.URL)
	lt.logger.Infof("Request Headers: %+v", req.Header)

	// mengganti request body karena sudah dibaca dalam fungsi logging
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

	// logging setelah menerima response
	lt.logger.Infof("Received response with status: %s", res.Status)
	lt.logger.Infof("Response Headers: %+v", res.Header)
	
	// menampilkan response body jika ada
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		lt.logger.Infof("Response Body: %+v", resBody)
	}
	res.Body = io.NopCloser(bytes.NewBuffer(resBody))
	return res, nil
}
