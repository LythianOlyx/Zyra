package observability

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.Logger
	loggerOnce   sync.Once
)

// InitLogger initializes the global Zap logger based on environment.
func InitLogger(env string) (*zap.Logger, error) {
	var err error
	loggerOnce.Do(func() {
		var cfg zap.Config
		if env == "production" {
			cfg = zap.NewProductionConfig()
			cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		} else {
			cfg = zap.NewDevelopmentConfig()
			cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}

		globalLogger, err = cfg.Build()
	})
	if err != nil {
		return nil, err
	}
	if globalLogger == nil {
		globalLogger = zap.NewNop()
	}
	return globalLogger, nil
}

// Logger returns the initialized global logger instance.
func Logger() *zap.Logger {
	if globalLogger == nil {
		l, _ := InitLogger("development")
		return l
	}
	return globalLogger
}

// LogAudit logs structured security and administrative audit events.
func LogAudit(event string, userID string, ip string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event_type", "security_audit"),
		zap.String("event", event),
		zap.String("user_id", userID),
		zap.String("ip", ip),
	}

	for k, v := range details {
		fields = append(fields, zap.Any(k, v))
	}

	Logger().Info("audit_event", fields...)
}
