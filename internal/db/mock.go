package db

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/mmiloslav/mock/pkg/stringtool"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Mock struct {
	ID      int    `gorm:"primaryKey"`
	Name    string `gorm:"not null"`
	Active  bool   `gorm:"not null"`
	GroupID int    `gorm:"not null"`
	Group   Group  `gorm:"not null;foreignKey:GroupID"`

	// RQ
	RqMethod      string `gorm:"not null"`
	RqPath        string `gorm:"not null"`
	RqBody        string `gorm:"type:text"`
	RqQueryParams datatypes.JSON

	// RS
	RsStatus  int `gorm:"not null"`
	RsHeaders datatypes.JSON
	RsBody    string `gorm:"type:text;not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

func GetMock(method, path, body string, queryParams map[string][]string) (Mock, error) {
	if stringtool.Empty(method) {
		return Mock{}, errors.New("method id empty")
	}

	if stringtool.Empty(path) {
		return Mock{}, errors.New("path id empty")
	}

	m := Mock{
		Active:   true,
		RqMethod: method,
		RqPath:   path,
	}

	tx := mockDB.Where(m)

	if len(queryParams) > 0 {
		jsonBytes, err := json.Marshal(queryParams)
		if err != nil {
			return Mock{}, err
		}

		str := string(jsonBytes)

		tx = tx.Where("JSON_CONTAINS(rq_query_params, ?) AND JSON_CONTAINS(?, rq_query_params)", str, str)
	}

	if body != "" {
		m.RqBody = body
	}

	err := tx.First(&m).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return Mock{}, err
		}

		return Mock{}, nil
	}

	return m, nil
}

func (m Mock) GetRsHeaders() (map[string][]string, error) {
	if len(m.RsHeaders) == 0 {
		return nil, nil
	}

	var result map[string][]string
	err := json.Unmarshal(m.RsHeaders, &result)

	return result, err
}

func (m *Mock) Create() error {
	return mockDB.Create(m).Error
}

func (m *Mock) Update() error {
	return mockDB.Save(m).Error
}

func (m *Mock) One() (bool, error) {
	err := mockDB.Where(m).First(&m).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return false, err
		}

		return false, nil
	}

	return true, nil
}

func MockExists(name string, groupID int) (bool, error) {
	var count int64
	err := mockDB.Model(&Mock{}).Where("name = ? AND group_id = ?", name, groupID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
