package gormtrace

import (
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"go.opentelemetry.io/otel"
	"gorm.io/gorm"

	"emshop/pkg/log"
)

// Enable wires the OpenTelemetry tracing plugin into the provided GORM DB.
// It is safe to call multiple times with the same *gorm.DB; subsequent calls
// are ignored by the underlying plugin registration.
func Enable(db *gorm.DB, dbName string) {
	if db == nil {
		return
	}

	opts := []otelgorm.Option{
		otelgorm.WithTracerProvider(otel.GetTracerProvider()),
	}
	if dbName != "" {
		opts = append(opts, otelgorm.WithDBName(dbName))
	}

	if err := db.Use(otelgorm.NewPlugin(opts...)); err != nil {
		log.Warnf("enable gorm tracing failed: %v", err)
	}
}
