package middlewares

import (
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    metric "emshop/gin-micro/core/metric"
)

// httpServerReqTotal is a CounterVec with labels: service, route, method, code
var httpServerReqTotal = metric.NewCounterVec(&metric.CounterVecOpts{
    Namespace: "http_server",
    Subsystem: "requests",
    Name:      "emshop_total",
    Help:      "http server requests count.",
    Labels:    []string{"service", "route", "method", "code"},
})

// HTTPMetricsMiddleware records per-request counters for QPS tracking.
// Use in combination with Prometheus scraping (/metrics) and rate() in Grafana.
func HTTPMetricsMiddleware(service string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Process request
        start := time.Now()
        _ = start // reserved for future duration metrics if needed
        c.Next()

        route := c.FullPath()
        if route == "" {
            route = "NOT_FOUND"
        }
        method := c.Request.Method
        code := strconv.Itoa(c.Writer.Status())

        // Increment request total counter
        httpServerReqTotal.Inc(service, route, method, code)
    }
}

