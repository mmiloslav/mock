package api

import (
	"github.com/mmiloslav/mock/internal/db"
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
