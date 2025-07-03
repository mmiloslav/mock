package db

import (
	"time"

	"gorm.io/gorm"
)

type Group struct {
	ID        int    `gorm:"primaryKey"`
	Name      string `gorm:"unique;not null"`
	Mocks     []Mock `gorm:"foreignKey:GroupID"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

func GetGroups(preloadMocks bool) ([]Group, error) {
	tx := mockDB
	if preloadMocks {
		tx = tx.Preload("Mocks")
	}

	var groups []Group
	err := tx.Order("name").Find(&groups).Error
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (m *Group) Create() error {
	return mockDB.Create(m).Error
}

func GroupExistsByName(name string) (bool, error) {
	group := Group{Name: name}

	return group.exists()
}

func GroupExistsByID(id int) (bool, error) {
	group := Group{ID: id}

	return group.exists()
}

func (m Group) exists() (bool, error) {
	err := mockDB.Where(m).First(&m).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
