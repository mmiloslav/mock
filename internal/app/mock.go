package app

import (
	"context"
	"io"
	"net/http"

	"encoding/json"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mmiloslav/mock/internal/db"
	"github.com/mmiloslav/mock/internal/myerrors"
	"github.com/mmiloslav/mock/internal/mylog"
	"github.com/mmiloslav/mock/pkg/stringtool"
)

const requestIDKey = "request_id"

// NewRouter creates mux.Router
func NewRouter() http.Handler {
	router := mux.NewRouter().StrictSlash(false)

	router.
		Methods([]string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}...).
		Path("/{rest:.*}").
		Name("Mock").
		Handler(requestIDMiddleware(http.HandlerFunc(mockHandler)))

	return http.Handler(router)
}

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

type baseRS struct {
	Success   bool   `json:"success"`
	ErrorCode string `json:"error_code"`
}

func (rs *baseRS) setError(err string) {
	rs.Success = false
	rs.ErrorCode = err
}

func mockHandler(w http.ResponseWriter, r *http.Request) {
	logger := mylog.Logger.WithField(requestIDKey, r.Context().Value(requestIDKey))
	logger.Info("mocking response...")

	rs := baseRS{}

	body := ""
	if r.Method != http.MethodGet {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Errorf("failed to read request body with error [%s]", err.Error())
			rs.setError(myerrors.ErrInternal)
			writeResponse(w, rs, http.StatusInternalServerError)
			return
		}

		body = string(bodyBytes)
	}

	mockDB, err := db.GetMock(r.Method, r.URL.Path, body, r.URL.Query())
	if err != nil {
		logger.Errorf("failed to find mock with error [%s]", err.Error())
		rs.setError(myerrors.ErrInternal)
		writeResponse(w, rs, http.StatusInternalServerError)
		return
	}
	if mockDB.ID == 0 {
		logger.Errorf("mock not found")
		rs.setError(myerrors.ErrNotFound)
		writeResponse(w, rs, http.StatusNotFound)
		return
	}

	headers, err := mockDB.GetRsHeaders()
	if err != nil {
		logger.Errorf("failed to get mock [%d] rs headers with error [%s]", mockDB.ID, err.Error())
		rs.setError(myerrors.ErrInternal)
		writeResponse(w, rs, http.StatusInternalServerError)
		return
	}

	for k, vals := range headers {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(mockDB.RsStatus)
	if !stringtool.Empty(mockDB.RsBody) {
		_, err := w.Write([]byte(mockDB.RsBody))
		if err != nil {
			logger.Errorf("failed to write rs body for mock [%d] with error [%s]", mockDB.ID, err.Error())
			rs.setError(myerrors.ErrInternal)
			writeResponse(w, rs, http.StatusInternalServerError)
			return
		}
	}
}
