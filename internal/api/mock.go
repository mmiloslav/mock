package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mmiloslav/mock/internal/db"
	"github.com/mmiloslav/mock/internal/myerrors"
	"github.com/mmiloslav/mock/internal/mylog"
	"github.com/mmiloslav/mock/pkg/maptool"
	"github.com/mmiloslav/mock/pkg/stringtool"
)

var validMethods = map[string]struct{}{
	http.MethodGet:     {},
	http.MethodHead:    {},
	http.MethodPost:    {},
	http.MethodPut:     {},
	http.MethodPatch:   {},
	http.MethodDelete:  {},
	http.MethodConnect: {},
	http.MethodOptions: {},
	http.MethodTrace:   {},
}

type getMocksRS struct {
	baseRS
	Groups []Group `json:"groups"`
}

func getMocksHandler(w http.ResponseWriter, r *http.Request) {
	logger := mylog.Logger.WithField(requestIDKey, r.Context().Value(requestIDKey))
	logger.Info("get mocks handler...")

	rs := getMocksRS{}

	dbGroups, err := db.GetGroups(true)
	if err != nil {
		logger.Errorf("failed to get groups & mocks with error [%s]", err.Error())
		rs.setError(myerrors.ErrInternal)
		writeResponse(w, rs, http.StatusInternalServerError)
		return
	}

	rs.Groups, err = newGroups(dbGroups)
	if err != nil {
		logger.Errorf("failed to convert groups with error [%s]", err.Error())
		rs.setError(myerrors.ErrInternal)
		writeResponse(w, rs, http.StatusInternalServerError)
		return
	}

	rs.setSuccess()
	writeResponse(w, rs, http.StatusOK)
}

type Mock struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`

	// RQ
	RqMethod      string                  `json:"rq_method"`
	RqPath        string                  `json:"rq_path"`
	RqBody        string                  `json:"rq_body,omitempty"`
	RqQueryParams []maptool.SortedJSONMap `json:"rq_query_params,omitempty"`

	// RS
	RsStatus  int                     `json:"rs_status"`
	RsHeaders []maptool.SortedJSONMap `json:"rs_headers,omitempty"`
	RsBody    string                  `json:"rs_body,omitempty"`
}

func newMocks(dbMocks []db.Mock) ([]Mock, error) {
	mocks := make([]Mock, 0, len(dbMocks))
	for _, dbMock := range dbMocks {
		mock, err := newMock(dbMock)
		if err != nil {
			return nil, err
		}

		mocks = append(mocks, mock)
	}

	return mocks, nil
}

func newMock(dbMock db.Mock) (Mock, error) {
	var queryParams map[string][]string
	if len(dbMock.RqQueryParams) > 0 {
		err := json.Unmarshal(dbMock.RqQueryParams, &queryParams)
		if err != nil {
			mylog.Logger.Errorf("failed to unmarshal query params for mock [%d]: [%s]", dbMock.ID, err)
			return Mock{}, err
		}
	}

	var rsHeaders map[string][]string
	if len(dbMock.RsHeaders) > 0 {
		err := json.Unmarshal(dbMock.RsHeaders, &rsHeaders)
		if err != nil {
			mylog.Logger.Errorf("failed to unmarshal response headers for mock [%d]: [%s]", dbMock.ID, err)
			return Mock{}, err
		}
	}

	return Mock{
		ID:            dbMock.ID,
		Name:          dbMock.Name,
		Active:        dbMock.Active,
		RqMethod:      dbMock.RqMethod,
		RqPath:        dbMock.RqPath,
		RqBody:        dbMock.RqBody,
		RqQueryParams: maptool.SortJSONMap(queryParams),
		RsStatus:      dbMock.RsStatus,
		RsHeaders:     maptool.SortJSONMap(rsHeaders),
		RsBody:        dbMock.RsBody,
	}, nil
}

type createMockRQ struct {
	Name    string `json:"name"`
	GroupID int    `json:"group_id"`

	//RQ
	RqMethod      string                  `json:"rq_method"`
	RqPath        string                  `json:"rq_path"`
	RqBody        string                  `json:"rq_body"`
	RqQueryParams []maptool.SortedJSONMap `json:"rq_query_params"`

	//RS
	RsStatus  int                     `json:"rs_status"`
	RsHeaders []maptool.SortedJSONMap `json:"rs_headers"`
	RsBody    string                  `json:"rs_body"`
}

func (rq createMockRQ) Validate() error {
	if stringtool.Empty(rq.Name) {
		return errors.New("name is empty")
	}

	if rq.GroupID <= 0 {
		return errors.New("groupID not valid")
	}

	// RQ
	if stringtool.Empty(rq.RqMethod) {
		return errors.New("rq method is empty")
	}

	if rq.RqMethod == http.MethodGet && !stringtool.Empty(rq.RqBody) {
		return errors.New("cannot add rq body for GET method")
	}

	if _, ok := validMethods[rq.RqMethod]; !ok {
		return errors.New("rq method is not valid")
	}

	if len(rq.RqQueryParams) > 0 {
		for _, qp := range rq.RqQueryParams {
			if stringtool.Empty(qp.Key) {
				return errors.New("query param is empty")
			}

			for _, v := range qp.Values {
				if stringtool.Empty(v) {
					return errors.New("query param value is empty")
				}
			}
		}
	}

	if stringtool.Empty(rq.RqPath) || !strings.HasPrefix(rq.RqPath, "/") {
		return errors.New("rq path is empty")
	}

	//RS
	if rq.RsStatus <= 0 {
		return errors.New("rs status not valid")
	}

	if len(rq.RsHeaders) > 0 {
		for _, h := range rq.RsHeaders {
			if stringtool.Empty(h.Key) {
				return errors.New("header is empty")
			}

			for _, v := range h.Values {
				if stringtool.Empty(v) {
					return errors.New("header value is empty")
				}
			}
		}
	}

	return nil
}

type createMockRS struct {
	baseRS
	ID int `json:"id"`
}

func createMockHandler(w http.ResponseWriter, r *http.Request) {
	logger := mylog.Logger.WithField(requestIDKey, r.Context().Value(requestIDKey))
	logger.Info("create mock handler...")

	rs := createMockRS{}
	rq := createMockRQ{}
	err := json.NewDecoder(r.Body).Decode(&rq)
	if err != nil {
		logger.Errorf("failed to decode request with error [%s]", err.Error())
		rs.setError(myerrors.ErrBadRequest)
		writeResponse(w, rs, http.StatusBadRequest)
		return
	}

	err = rq.Validate()
	if err != nil {
		logger.Errorf("request is not valid: [%s]", err.Error())
		rs.setError(myerrors.ErrBadRequest)
		writeResponse(w, rs, http.StatusBadRequest)
		return
	}

	ok, err := db.GroupExistsByID(rq.GroupID)
	if err != nil {
		logger.Errorf("failed to check if group exists with error [%s]", err.Error())
		rs.setError(myerrors.ErrInternal)
		writeResponse(w, rs, http.StatusInternalServerError)
		return
	}
	if !ok {
		logger.Errorf("group with id [%d] does not exist", rq.GroupID)
		rs.setError(myerrors.ErrGroupNotExists)
		writeResponse(w, rs, http.StatusConflict)
		return
	}

	ok, err = db.MockExists(rq.Name, rq.GroupID)
	if err != nil {
		logger.Errorf("failed to check if mock exists with error [%s]", err.Error())
		rs.setError(myerrors.ErrInternal)
		writeResponse(w, rs, http.StatusInternalServerError)
		return
	}
	if ok {
		logger.Errorf("mock with name [%s] already exists in group [%d]", rq.Name, rq.GroupID)
		rs.setError(myerrors.ErrMockNameExists)
		writeResponse(w, rs, http.StatusConflict)
		return
	}

	queryParams, err := json.Marshal(maptool.UnsortJSONMap(rq.RqQueryParams))
	if err != nil {
		logger.Errorf("failed to marshal query params with error [%s]", err.Error())
		rs.setError(myerrors.ErrInternal)
		writeResponse(w, rs, http.StatusInternalServerError)
		return
	}

	headers, err := json.Marshal(maptool.UnsortJSONMap(rq.RsHeaders))
	if err != nil {
		logger.Errorf("failed to marshal headers with error [%s]", err.Error())
		rs.setError(myerrors.ErrInternal)
		writeResponse(w, rs, http.StatusInternalServerError)
		return
	}

	mock := db.Mock{
		Name:          rq.Name,
		Active:        true,
		GroupID:       rq.GroupID,
		RqMethod:      rq.RqMethod,
		RqPath:        rq.RqPath,
		RqBody:        rq.RqBody,
		RqQueryParams: queryParams,
		RsStatus:      rq.RsStatus,
		RsHeaders:     headers,
		RsBody:        rq.RsBody,
	}
	err = mock.Create()
	if err != nil {
		logger.Errorf("failed to create mock with error [%s]", err.Error())
		rs.setError(myerrors.ErrInternal)
		writeResponse(w, rs, http.StatusInternalServerError)
		return
	}

	rs.ID = mock.ID
	rs.setSuccess()
	writeResponse(w, rs, http.StatusCreated)
}

func activateMockHandler(w http.ResponseWriter, r *http.Request) {
	logger := mylog.Logger.WithField(requestIDKey, r.Context().Value(requestIDKey))
	logger.Info("activate mock handler...")

	rs := baseRS{}

	vars := mux.Vars(r)
	mockID, err := strconv.Atoi(vars["mock_id"])
	if err != nil {
		logger.Errorf("failed to convert mock_id with error [%s]", err.Error())
		rs.setError(myerrors.ErrBadRequest)
		writeResponse(w, rs, http.StatusBadRequest)
		return
	}

	mockDB := db.Mock{ID: mockID}
	ok, err := mockDB.One()
	if err != nil {
		logger.Errorf("failed to check if mock exists with error [%s]", err.Error())
		rs.setError(myerrors.ErrInternal)
		writeResponse(w, rs, http.StatusInternalServerError)
		return
	}
	if !ok {
		logger.Errorf("mock with id [%d] does not exist", mockID)
		rs.setError(myerrors.ErrMockNotExists)
		writeResponse(w, rs, http.StatusConflict)
		return
	}

	mockDB.Active = !mockDB.Active
	err = mockDB.Update()
	if err != nil {
		logger.Errorf("failed to activate mock with error [%s]", err.Error())
		rs.setError(myerrors.ErrInternal)
		writeResponse(w, rs, http.StatusInternalServerError)
		return
	}

	rs.setSuccess()
	writeResponse(w, rs, http.StatusOK)
}
