package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mmiloslav/mock/internal/db"
	"github.com/mmiloslav/mock/internal/myerrors"
	"github.com/mmiloslav/mock/internal/mylog"
	"github.com/mmiloslav/mock/pkg/stringtool"
)

type Group struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Mocks []Mock `json:"mocks,omitempty"`
}

func newGroups(dbGroups []db.Group) ([]Group, error) {
	groups := make([]Group, 0, len(dbGroups))
	for _, dbGroups := range dbGroups {
		group, err := newGroup(dbGroups)
		if err != nil {
			return nil, err
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func newGroup(dbGroup db.Group) (Group, error) {
	mocks, err := newMocks(dbGroup.Mocks)
	if err != nil {
		return Group{}, err
	}

	return Group{
		ID:    dbGroup.ID,
		Name:  dbGroup.Name,
		Mocks: mocks,
	}, nil
}

type createGroupRQ struct {
	Name string `json:"name"`
}

func (rq createGroupRQ) Validate() error {
	if stringtool.Empty(rq.Name) {
		return errors.New("name is empty")
	}

	return nil
}

type createGroupRS struct {
	baseRS
	ID int `json:"id"`
}

func createGroupHandler(w http.ResponseWriter, r *http.Request) {
	logger := mylog.Logger.WithField(requestIDKey, r.Context().Value(requestIDKey))
	logger.Info("create group handler...")

	rs := createGroupRS{}
	rq := createGroupRQ{}
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

	ok, err := db.GroupExistsByName(rq.Name)
	if err != nil {
		logger.Errorf("failed to check if group already exists with error [%s]", err.Error())
		rs.setError(myerrors.ErrInternal)
		writeResponse(w, rs, http.StatusInternalServerError)
		return
	}
	if ok {
		logger.Errorf("group with name [%s] already exists", rq.Name)
		rs.setError(myerrors.ErrGroupAlreadyExists)
		writeResponse(w, rs, http.StatusConflict)
		return
	}

	group := db.Group{Name: rq.Name}
	err = group.Create()
	if err != nil {
		logger.Errorf("failed to create group with error [%s]", err.Error())
		rs.setError(myerrors.ErrInternal)
		writeResponse(w, rs, http.StatusInternalServerError)
		return
	}

	rs.ID = group.ID
	rs.setSuccess()
	writeResponse(w, rs, http.StatusCreated)
}

type getGroupsRS struct {
	baseRS
	Groups []Group `json:"groups"`
}

func getGroupsHandler(w http.ResponseWriter, r *http.Request) {
	logger := mylog.Logger.WithField(requestIDKey, r.Context().Value(requestIDKey))
	logger.Info("get groups handler...")

	rs := getGroupsRS{}

	dbGroups, err := db.GetGroups(false)
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
