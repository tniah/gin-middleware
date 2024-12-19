package auditlogger

import (
	"github.com/gin-gonic/gin"
	"time"
)

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
	// LogURI instructs logger to extract request URI (i.e. `/api/v1/users?name=kai`
	LogURI bool
	// LogURIPath instructs logger to extract request URI path (i.e. `/api/v1/users`)
	LogURIPath bool
	// LogRequestID instructs logger to extract request ID from request `X-Request-ID` header.
	LogRequestID bool
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
	var skipPaths map[string]bool
	if length := len(cfg.SkipPaths); length > 0 {
		skipPaths = make(map[string]bool, length)
		for _, path := range cfg.SkipPaths {
			skipPaths[path] = true
		}
	}

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

		if cfg.LogResponseSize {
			params.ResponseSize = c.Writer.Size()
		}

		if cfg.LogLatency {
			params.Latency = time.Since(startTime)
		}

		if cfg.LogValuesFunc != nil {
			cfg.LogValuesFunc(c, params)
		}
	}
}
