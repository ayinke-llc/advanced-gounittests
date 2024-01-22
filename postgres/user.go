package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type User struct {
	ID       uuid.UUID
	Email    string
	FullName string
}

type userRepo struct {
	inner *sql.DB
}

func NewUserRepository(db *sql.DB) *userRepo {
	return &userRepo{
		inner: db,
	}
}

func (u *userRepo) Get(ctx context.Context, email string) (*User, error) {
	sqlStatement := `SELECT id,email,full_name FROM users WHERE email=$1;`
	user := new(User)
	row := u.inner.QueryRow(sqlStatement, email)

	return user, row.Scan(&user.ID, &user.Email, &user.FullName)
}

func (u *userRepo) Create(ctx context.Context, user *User) error {
	sqlStatement := `
INSERT INTO users (email, full_name)
VALUES ($1, $2)`
	_, err := u.inner.Exec(sqlStatement, user.Email, user.FullName)
	return err
}
