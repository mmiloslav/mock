package api

import (
	"context"
	"io"
	"net/http"

	"encoding/json"

	"github.com/google/uuid"
	"github.com/mmiloslav/mock/internal/db"
	"github.com/mmiloslav/mock/internal/myerrors"
	"github.com/mmiloslav/mock/internal/mylog"

	"github.com/gorilla/mux"
)

const requestIDKey = "request_id"

type route struct {
	Name               string
	Method             string
	Pattern            string
	HandlerFunc        http.HandlerFunc
	MiddlewareAuthFunc func(http.Handler) http.Handler
}

// API
var routes = []route{
	// PING
	{Name: "Ping", Method: http.MethodGet, Pattern: "/api/ping", HandlerFunc: pingHandler, MiddlewareAuthFunc: requestIDMiddleware},

	// MOCK
	{Name: "Get Mocks", Method: http.MethodGet, Pattern: "/api/v1/mocks", HandlerFunc: getMocksHandler, MiddlewareAuthFunc: requestIDMiddleware},
	{Name: "Create Mock", Method: http.MethodPost, Pattern: "/api/v1/mocks", HandlerFunc: createMockHandler, MiddlewareAuthFunc: requestIDMiddleware},

	// GROUP
	{Name: "Get Groups", Method: http.MethodGet, Pattern: "/api/v1/groups", HandlerFunc: getGroupsHandler, MiddlewareAuthFunc: requestIDMiddleware},
	{Name: "Create Group", Method: http.MethodPost, Pattern: "/api/v1/groups", HandlerFunc: createGroupHandler, MiddlewareAuthFunc: requestIDMiddleware},
}

// newRouter creates mux.Router
func NewRouter() http.Handler {
	router := mux.NewRouter().StrictSlash(false)
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.MiddlewareAuthFunc(route.HandlerFunc))
	}

	return http.Handler(router)
}

// requestIDMiddleware
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		w.Header().Set("X-Request-ID", requestID)

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// writeResponse writes http response
func writeResponse(w http.ResponseWriter, r interface{}, statusCode int) {
	byteBody, err := json.Marshal(r)
	if err != nil {
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	io.Writer.Write(w, byteBody)
}

// BASE RESPONSE

type baseRS struct {
	Success   bool   `json:"success"`
	ErrorCode string `json:"error_code"`
}

func (rs *baseRS) setSuccess() {
	rs.Success = true
	rs.ErrorCode = ""
}

func (rs *baseRS) setError(err string) {
	rs.Success = false
	rs.ErrorCode = err
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	logger := mylog.Logger.WithField("handler", "ping")

	rs := baseRS{}

	err := db.Ping()
	if err != nil {
		logger.Errorf("failed to ping db with error [%s]", err.Error())

		rs.setError(myerrors.ErrInternal)
		writeResponse(w, rs, http.StatusInternalServerError)
		return
	}

	rs.setSuccess()
	writeResponse(w, rs, http.StatusOK)
}
