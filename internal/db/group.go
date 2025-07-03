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

func GetGroups() ([]Group, error) {
	var groups []Group
	err := mockDB.Preload("Mocks").Order("name").Find(&groups).Error
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (m *Group) Create() error {
	return mockDB.Create(m).Error
}

func GroupExists(name string) (bool, error) {
	group := Group{Name: name}
	err := mockDB.Where(group).First(&group).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
