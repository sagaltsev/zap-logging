package logging

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logLevel = map[string]zapcore.Level{
		"debug":  zapcore.DebugLevel,
		"info":   zapcore.InfoLevel,
		"warn":   zapcore.WarnLevel,
		"error":  zapcore.ErrorLevel,
		"dpanic": zapcore.DPanicLevel,
		"panic":  zapcore.PanicLevel,
		"fatal":  zapcore.FatalLevel,
	}

	// logger is a regular zap logger
	logger = &zap.Logger{}
	// Logger is a sugared zap logger
	Logger = &zap.SugaredLogger{}
)

const (
	// Info logger level
	Info = "info"
	// Debug logger level
	Debug = "debug"
	// Error logger level
	Error = "error"
	// Warn logger level
	Warn = "warn"
	// Panic logger level
	Panic = "panic"
	// Fatal logger level
	Fatal = "fatal"
)

func init() {
	Set("info")
}

// Set setting Logger and internal logger with log level
func Set(level string) {
	l, _ := zap.Config{
		Encoding:    "json",
		Level:       zap.NewAtomicLevelAt(logLevel[strings.ToLower(level)]),
		OutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  "message",
			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,
			TimeKey:     "time",
			EncodeTime:  zapcore.ISO8601TimeEncoder,
		},
	}.Build()

	logger = l
	Logger = l.Sugar()
}

// LogAppStart logs application start event
func LogAppStart(name string, args interface{}) {
	var fields map[string]interface{}

	jsonString, err := json.Marshal(args)
	if err == nil {
		err := json.Unmarshal(jsonString, &fields)
		if err != nil {
			fields["parse_error"] = err.Error()
		}
	}

	typeField := zap.String("type", "lifecycle")
	eventField := zap.String("event", "start")

	zapFields := make([]zapcore.Field, 0)

	zapFields = append(zapFields, typeField)
	zapFields = append(zapFields, eventField)

	for k, v := range fields {
		zapField := zap.Any(k, v)
		zapFields = append(zapFields, zapField)
	}

	message := fmt.Sprintf("starting application: %v", name)
	logger.Info(message, zapFields...)
}

// LogAppStop logs application stop event
func LogAppStop(name string, signal os.Signal, err error) {
	typeField := zap.String("type", "lifecycle")
	eventField := zap.String("event", "stop")

	if err != nil {
		message := fmt.Sprintf("stopping application: %v (%v)", name, err)
		logger.Error(message, typeField, eventField)
	} else {
		message := fmt.Sprintf("stopping application: %v (%v)", name, signal)
		logger.Info(message, typeField, eventField)
	}
}

// LogRequest logs request event
func LogRequest(r *http.Request, statusCode int) {
	logger.Info(strconv.Itoa(statusCode) + " ->" + r.Method + " " + r.RequestURI,
		zap.String("type", "access"),
		zap.String("event", "request"),
		zap.String("remote_ip", getRemoteIP(r)),
		zap.String("host", r.Host),
		zap.String("url_path", r.RequestURI),
		zap.String("method", r.Method),
		zap.Int("status_code", statusCode),
		zap.String("correlation_id", r.Header.Get(CorrelationIDHeader)),
		zap.String("user_correlation_id", r.Header.Get(UserCorrelationIDHeader)),
	)
}

// LogWithCorrelationIDS logs message with provided severity and correlation IDS
func LogWithCorrelationIDS(severity string, message string, correlationID string, userCorrelationID string) {
	lg := logger.With(
		zap.String("correlation_id", correlationID),
		zap.String("user_correlation_id", userCorrelationID),
	)
	switch severity {
	case Debug:
		lg.Debug(message)
	case Error:
		lg.Error(message)
	case Warn:
		lg.Warn(message)
	case Fatal:
		lg.Fatal(message)
	case Panic:
		lg.Panic(message)
	default:
		lg.Info(message)
	}
}

func getRemoteIP(r *http.Request) string {
	if r.Header.Get("X-Cluster-Client-Ip") != "" {
		return r.Header.Get("X-Cluster-Client-Ip")
	}

	if r.Header.Get("X-Real-Ip") != "" {
		return r.Header.Get("X-Real-Ip")
	}

	return strings.Split(r.RemoteAddr, ":")[0]
}
