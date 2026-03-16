package repository

import (
	"database-exercise/internal/model"

	"github.com/jmoiron/sqlx"
)

type UserSqlxRepository struct {
	db *sqlx.DB
}

func (r *UserSqlxRepository) Create(user *model.User) error {
	query := `INSERT INTO users (name, email) VALUES($1, $2) RETURNING id`
	return r.db.QueryRow(query, user.Name, user.Email).Scan(&user.ID)
}

func (r *UserSqlxRepository) GetByID(id int64) (*model.User, error) {
	var user model.User
	query := `SELECT id, name, email, created_at FROM users WHERE id = $1`
	err := r.db.Get(&user, query, id)
	return &user, err
}

func (r *UserSqlxRepository) List() ([]model.User, error) {
	var users []model.User
	query := `SELECT id, name, email, created_at FROM users`
	err := r.db.Select(&users, query)
	return users, err
}

func (r *UserSqlxRepository) Delete(id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func NewUserSqlxRepository(db *sqlx.DB) *UserSqlxRepository {
	return &UserSqlxRepository{db: db}
}
