package repository

import (
	"agenda-api/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (id, email, password, name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowx(
		query,
		user.ID, user.Email, user.Password, user.Name, user.Role, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE id = $1`
	err := r.db.Get(&user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE email = $1`
	err := r.db.Get(&user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := r.db.Get(&exists, query, email)
	return exists, err
}

func (r *UserRepository) GetAll() ([]models.User, error) {
	var users []models.User
	query := `SELECT * FROM users ORDER BY created_at DESC`
	err := r.db.Select(&users, query)
	return users, err
}

func (r *UserRepository) GetByRole(role models.Role) ([]models.User, error) {
	var users []models.User
	query := `SELECT * FROM users WHERE role = $1 ORDER BY name`
	err := r.db.Select(&users, query, role)
	return users, err
}
