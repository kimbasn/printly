package service

import (
	"time"

	"github.com/kimbasn/printly/internal/entity"

	"github.com/kimbasn/printly/internal/repository"
)

type UserService interface {
	RegisterIfNotExist(user *entity.User) (*entity.User, error)
	GetByUID(uid string) (*entity.User, error)
	DeleteByUID(uid string) error
	UpdateProfile(user *entity.User) error
	FindAll() []entity.User
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) RegisterIfNotExist(u *entity.User) (*entity.User, error) {
	existing, err := s.repo.FindByUID(u.UID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	if err := s.repo.Create(u); err != nil {
		return nil, err
	}
	return u, nil

}

func (s *userService) GetByUID(uid string) (*entity.User, error) {
	return s.repo.FindByUID(uid)
}

func (s *userService) UpdateProfile(u *entity.User) error {
	u.UpdatedAt = time.Now()
	return s.repo.Update(u)
}

func (s *userService) FindAll() []entity.User {
	return s.repo.FindAll()
}

func (s *userService) DeleteByUID(uid string) error {
	return s.repo.DeleteByUID(uid)
}