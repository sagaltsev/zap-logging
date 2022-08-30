package logging

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

const (
	// CorrelationIDHeader header key
	CorrelationIDHeader = "X-Correlation-Id"
	// UserCorrelationIDHeader header key
	UserCorrelationIDHeader = "X-User-Correlation-Id"
)

// HTTPHandlerMiddleware middleware for standard http.Handler
func HTTPHandlerMiddleware(next http.Handler) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		lrw := NewLoggingResponseWriter(rw)
		next.ServeHTTP(lrw, r)
		
		LogRequest(r, lrw.StatusCode)
	}
}

// HTTPRouterMiddleware middleware for httprouter: https://github.com/julienschmidt/httprouter
func HTTPRouterMiddleware(next httprouter.Handle) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		lrw := NewLoggingResponseWriter(rw)
		next(lrw, r, ps)

		LogRequest(r, lrw.StatusCode)
	}
}

// CorrelationMiddleware adds correlationID and userCorrelationID to handler
func CorrelationMiddleware(next http.Handler) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(CorrelationIDHeader)
		if id == "" {
			id = uuid.NewString()
			r.Header.Set(CorrelationIDHeader, id)
		}

		userID := r.Header.Get(UserCorrelationIDHeader)
		if userID == "" {
			userID = uuid.NewString()
			r.Header.Set(UserCorrelationIDHeader, userID)
		}

		rw.Header().Set(CorrelationIDHeader, id)
		rw.Header().Set(UserCorrelationIDHeader, userID)

		next.ServeHTTP(rw, r)
	}
}

// ResponseWriter object
type ResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

// NewLoggingResponseWriter constructor
func NewLoggingResponseWriter(rw http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{rw, http.StatusOK}
}

// WriteHeader decorates response header with status code
func (lrw *ResponseWriter) WriteHeader(code int) {
	lrw.StatusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
