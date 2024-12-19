package auditlogger

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

const HeaderContentLength = "Content-Length"

type Skipper func(c *gin.Context) bool

type LoggerConfig struct {
	// Skipper defines a function to skip middleware
	Skipper Skipper

	// SkipPaths is an url path array which logs are not written.
	SkipPaths []string

	// LogValuesFunc defines a function that is called with values extracted by logger.
	LogValuesFunc func(c *gin.Context, v RequestLoggerParams)

	// LogLatency instructs logger to record how much time the server cost to process a certain request.
	LogLatency bool
	// LogProtocol instructs logger to extract request protocol (i.e. `HTTP/1.1` or `HTTP/2`)
	LogProtocol bool
	// LogRemoteIP instructs logger to extract request remote IP. It equals Context's ClientIP method.
	LogRemoteIP bool
	// LogHost instructs logger to extract request host value (i.e. `example.com`)
	LogHost bool
	// LogMethod instructs logger to extract
	LogMethod bool
	// LogURI instructs logger to extract request URI (i.e. `/api/v1/users?name=makai`
	LogURI bool
	// LogURIPath instructs logger to extract request URI path (i.e. `/api/v1/users`)
	LogURIPath bool
	// LogRoutePath bool instructs logger to extract route path part to which request was matched
	// (i.e. `/api/v1/users/:userID`)
	LogRoutePath bool
	// LogRequestID instructs logger to extract request ID from one of the given parameters.
	LogRequestIdParams []string
	// LogReferer instructs logger to extract request referer values.
	LogReferer bool
	// LogUserAgent instructs logger to extract request user agent value.
	LogUserAgent bool
	// LogStatus instructs logger to extract HTTP response code.
	LogStatus bool
	// LogError instructs logger to extract error returned from executed handler chain.
	LogError bool
	// LogContentLength instructs logger to extract content length header value.
	LogContentLength bool
	// LogResponseSize instructs logger to extract response content length value.
	LogResponseSize bool
	// LogHeaders instructs logger to extract given list of headers from request.
	LogHeaders []string
	// LogQueryParams instructs logger to extract given list of query parameters from request.
	LogQueryParams []string
}

type RequestLoggerParams struct {
	StartTime     time.Time
	Latency       time.Duration
	Protocol      string
	RemoteIP      string
	Host          string
	Method        string
	URI           string
	URIPath       string
	RoutePath     string
	RequestID     string
	Referer       string
	UserAgent     string
	Status        int
	Error         string
	ContentLength string
	ResponseSize  int
	Headers       map[string][]string
	QueryParams   map[string][]string
}

func LoggerWithConfig(cfg LoggerConfig) gin.HandlerFunc {
	mw, err := cfg.ToMiddleware()
	if err != nil {
		panic(err)
	}
	return mw
}

func (cfg *LoggerConfig) ToMiddleware() (gin.HandlerFunc, error) {
	var skipPaths map[string]bool
	if length := len(cfg.SkipPaths); length > 0 {
		skipPaths = make(map[string]bool, length)
		for _, path := range cfg.SkipPaths {
			skipPaths[path] = true
		}
	}

	if cfg.LogValuesFunc == nil {
		return nil, errors.New("missing LogValuesFunc callback function for audit logger middleware")
	}

	logHeaders := len(cfg.LogHeaders) > 0
	headers := append([]string(nil), cfg.LogHeaders...)
	for i, v := range headers {
		headers[i] = http.CanonicalHeaderKey(v)
	}

	logQueryParams := len(cfg.LogQueryParams) > 0
	logRequestID := len(cfg.LogRequestIdParams) > 0

	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Process request
		c.Next()

		path := c.Request.URL.Path
		if ok := skipPaths[path]; ok || (cfg.Skipper != nil && cfg.Skipper(c)) {
			return
		}

		params := RequestLoggerParams{StartTime: startTime}
		if cfg.LogProtocol {
			params.Protocol = c.Request.Proto
		}

		if cfg.LogRemoteIP {
			params.RemoteIP = c.ClientIP()
		}

		if cfg.LogHost {
			params.Host = c.Request.Host
		}

		if cfg.LogMethod {
			params.Method = c.Request.Method
		}

		if cfg.LogURI {
			params.URI = c.Request.RequestURI
		}

		if cfg.LogURIPath {
			params.URIPath = c.Request.URL.Path
		}

		if cfg.LogRoutePath {
			params.RoutePath = c.FullPath()
		}

		if logRequestID {
			var reqID string
			for _, param := range cfg.LogRequestIdParams {
				reqID = c.GetHeader(param)
				if reqID == "" {
					reqID = c.Writer.Header().Get(param)
				}

				if reqID != "" {
					break
				}
			}
			params.RequestID = reqID
		}

		if cfg.LogReferer {
			params.Referer = c.Request.Referer()
		}

		if cfg.LogUserAgent {
			params.UserAgent = c.Request.UserAgent()
		}

		if cfg.LogStatus {
			params.Status = c.Writer.Status()
		}

		if cfg.LogError {
			params.Error = c.Errors.ByType(gin.ErrorTypePrivate).String()
		}

		if cfg.LogContentLength {
			params.ContentLength = c.Request.Header.Get(HeaderContentLength)
		}

		if cfg.LogResponseSize {
			params.ResponseSize = c.Writer.Size()
		}

		if logHeaders {
			params.Headers = map[string][]string{}
			for _, header := range headers {
				if values, ok := c.Request.Header[header]; ok {
					params.Headers[header] = values
				}
			}
		}

		if logQueryParams {
			params.QueryParams = map[string][]string{}
			for _, param := range cfg.LogQueryParams {
				if values, ok := c.GetQueryArray(param); ok {
					params.QueryParams[param] = values
				}
			}
		}

		if cfg.LogLatency {
			params.Latency = time.Since(startTime)
		}

		cfg.LogValuesFunc(c, params)
	}, nil
}
