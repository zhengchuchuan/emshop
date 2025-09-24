package middlewares

import (
    "fmt"
    "strings"
    "time"

    "github.com/alibaba/sentinel-golang/api"
    "github.com/alibaba/sentinel-golang/core/base"
    "github.com/gin-gonic/gin"
)

// HTTPConfig configures the Sentinel Gin middleware.
type HTTPConfig struct {
    // ResourcePrefix is prefixed to the resource name (e.g. service name).
    ResourcePrefix string
    // IncludePath if true, include the request path as resource.
    IncludePath bool
    // IncludeMethod if true, include the HTTP method.
    IncludeMethod bool
    // BlockFallback handles blocked request; if nil returns 429 with JSON.
    BlockFallback func(c *gin.Context, blockErr error)
    // ShouldProtect decides whether to apply Sentinel to this HTTP request.
    // If nil, protection is applied to all requests.
    ShouldProtect func(c *gin.Context) bool
}

// DefaultHTTPConfig creates a default config using serviceName as prefix.
func DefaultHTTPConfig(serviceName string) *HTTPConfig {
    return &HTTPConfig{
        ResourcePrefix: strings.ReplaceAll(serviceName, " ", "-"),
        IncludePath:    true,
        IncludeMethod:  true,
    }
}

// Sentinel returns a Gin middleware that protects HTTP endpoints with Sentinel.
func Sentinel(cfg *HTTPConfig) gin.HandlerFunc {
    if cfg == nil {
        cfg = &HTTPConfig{IncludePath: true, IncludeMethod: true}
    }
    return func(c *gin.Context) {
        if cfg.ShouldProtect != nil && !cfg.ShouldProtect(c) {
            c.Next()
            return
        }
        start := time.Now()
        resource := buildHTTPResource(cfg, c)
        entry, err := api.Entry(resource, api.WithTrafficType(base.Inbound))
        if err != nil {
            if cfg.BlockFallback != nil {
                cfg.BlockFallback(c, err)
                return
            }
            // default 429
            c.AbortWithStatusJSON(429, gin.H{
                "code": 429,
                "message": "Too Many Requests",
            })
            return
        }
        // proceed
        c.Next()
        // record handler error if any
        if len(c.Errors) > 0 {
            api.TraceError(entry, fmt.Errorf(c.Errors.String()))
        }
        // exit and finalize
        entry.Exit()
        _ = start // reserved for RT metrics if needed later
    }
}

func buildHTTPResource(cfg *HTTPConfig, c *gin.Context) string {
    var b strings.Builder
    if cfg.ResourcePrefix != "" {
        b.WriteString(cfg.ResourcePrefix)
        b.WriteString(":")
    }
    if cfg.IncludeMethod {
        b.WriteString(c.Request.Method)
    }
    if cfg.IncludePath {
        if cfg.IncludeMethod {
            b.WriteString(" ")
        }
        // use full path (e.g. /api/v1/users/:id) if available
        path := c.FullPath()
        if path == "" { // not a registered route
            path = c.Request.URL.Path
        }
        b.WriteString(path)
    }
    if b.Len() == 0 {
        return c.Request.URL.Path
    }
    return b.String()
}
