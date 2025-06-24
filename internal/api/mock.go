package api

import (
	"encoding/json"
	"net/http"

	"github.com/mmiloslav/mock/internal/db"
	"github.com/mmiloslav/mock/internal/myerrors"
	"github.com/mmiloslav/mock/internal/mylog"
	"github.com/mmiloslav/mock/pkg/maptool"
)

type getMocksRS struct {
	baseRS
	Groups []Group `json:"groups"`
}

func getMocksHandler(w http.ResponseWriter, r *http.Request) {
	logger := mylog.Logger.WithField(requestIDKey, r.Context().Value(requestIDKey))
	logger.Info("get mocks handler...")

	rs := getMocksRS{}

	dbGroups, err := db.GetGroups()
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
