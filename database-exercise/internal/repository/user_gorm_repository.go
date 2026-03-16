package repository

import (
	"database-exercise/internal/model"

	"gorm.io/gorm"
)

type UserGormRepository struct {
	db *gorm.DB
}

func (r *UserGormRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserGormRepository) GetByID(id int64) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *UserGormRepository) List() ([]model.User, error) {
	var users []model.User
	err := r.db.Find(&users).Error
	return users, err
}

func (r *UserGormRepository) Delete(id int64) error {
	return r.db.Delete(&model.User{}, id).Error
}

func NewUserGormRepository(db *gorm.DB) *UserGormRepository {
	return &UserGormRepository{db: db}
}
