package logging

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func TestHTTPHandlerMiddleware(t *testing.T) {
	m := HTTPHandlerMiddleware(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))

	rw := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "http://testurl.com", nil)
	r.RequestURI = "/"

	output := captureOutput(func() {
		m.ServeHTTP(rw, r)
	})

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Contains(t, output, `200 ->GET /`)
	assert.Contains(t, output, `"type": "access"`)
	assert.Contains(t, output, `"event": "request"`)
	assert.Contains(t, output, `"method": "GET"`)
	assert.Contains(t, output, `"status_code": 200`)
}

func TestHTTPRouterMiddleware(t *testing.T) {
	m := HTTPRouterMiddleware(httprouter.Handle(func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		rw.WriteHeader(http.StatusNotFound)
	}))

	rw := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "http://testurl.com", nil)
	r.RequestURI = "/"

	output := captureOutput(func() {
		m(rw, r, httprouter.Params{})
	})

	assert.Equal(t, http.StatusNotFound, rw.Code)
	assert.Contains(t, output, `404 ->GET /`)
	assert.Contains(t, output, `"type": "access"`)
	assert.Contains(t, output, `"event": "request"`)
	assert.Contains(t, output, `"method": "GET"`)
	assert.Contains(t, output, `"status_code": 404`)
}

func Test_CorrelationMiddleware(t *testing.T) {
	m := CorrelationMiddleware(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))

	rw := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "http://testurl.com", nil)

	_ = captureOutput(func() {
		m.ServeHTTP(rw, r)
	})

	assert.Equal(t, http.StatusOK, rw.Code)
}
