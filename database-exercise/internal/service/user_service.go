package service

import (
	"database-exercise/internal/model"
	"database-exercise/internal/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func (s *UserService) CreateUser(name, email string) (*model.User, error) {
	user := &model.User{
		Name:  name,
		Email: email,
	}
	err := s.repo.Create(user)
	return user, err
}

func (s *UserService) GetUserByID(id int64) (*model.User, error) {
	return s.repo.GetByID(id)
}

func (s *UserService) ListUsers() ([]model.User, error) {
	return s.repo.List()
}

func (s *UserService) DeleteUser(id int64) error {
	return s.repo.Delete(id)
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}
