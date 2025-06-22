package repository

import (
	"errors"
	"github.com/kimbasn/printly/internal/entity"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *entity.User) error
	FindByUID(uid string) (*entity.User, error)
	DeleteByUID(uid string) error
	Update(user *entity.User) error
	FindAll() []entity.User
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(u *entity.User) error {
	return r.db.Create(u).Error
}

func (r *userRepository) FindByUID(uid string) (*entity.User, error) {
	var u entity.User
	result := r.db.First(&u, "uid = ?", uid)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, result.Error
}

func (r *userRepository) Update(u *entity.User) error {
	return r.db.Save(u).Error
}

func (r *userRepository) FindAll() []entity.User {
	var users []entity.User
	r.db.Set("gorm:auto_preload", true).Find(&users)
	return  users
}

func (r *userRepository) DeleteByUID(uid string) error {
	return r.db.Where("uuid = ?", uid).Delete(&entity.User{}).Error
}