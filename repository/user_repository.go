package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/infrastructure/pqsql"
)

type useRepository struct {
	database pqsql.Client
}

// GetMailOrPhone implements domain.UserRepository.
func (ur *useRepository) GetMailOrPhone(ctx context.Context, email_bidx string, phone_bidx string, fn func(data *domain.User)) (*domain.User, error) {
	var user domain.User
	if fn != nil {
		fn(&user)
	}

	query := `SELECT id, email, phone FROM users WHERE email_bidx = $1 OR phone_bidx = $2 LIMIT 1`
	err := ur.database.Database().QueryRowContext(ctx, query, email_bidx, phone_bidx).Scan(
		&user.ID, &user.Email, &user.Phone,
	)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, errors.New("user not found")
	case err != nil:
		return nil, err
	}

	return &user, nil
}

func (ur *useRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (email, email_bidx, phone, phone_bidx, password_hash) VALUES ($1, $2, $3, $4, $5)`

	_, err := ur.database.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) (any, error) {
		_, err := tx.ExecContext(ctx, query, user.Email, user.EmailBidx, user.Phone, user.PhoneBidx, user.PasswordHash)
		return nil, err
	})

	if err != nil {
		fmt.Println("Error inserting user:", err)
		return err
	}

	return nil
}

func (ur *useRepository) GetUserByPhone(ctx context.Context, phone_bidx string, fn func(data *domain.User)) (*domain.User, error) {
	var user domain.User

	if fn != nil {
		fn(&user)
	}

	query := `SELECT id, email, phone FROM users WHERE phone_bidx = $1 LIMIT 1`
	err := ur.database.Database().QueryRowContext(ctx, query, nil, &domain.User{}, phone_bidx).Scan(
		&user.ID, &user.Email, &user.Phone, &user.Phone,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (ur *useRepository) GetUserByEmail(ctx context.Context, email_bidx string, fn func(data *domain.User)) (*domain.User, error) {
	var existingUser domain.User

	if fn != nil {
		fn(&existingUser)
	}

	query := `SELECT id, email, phone FROM users WHERE email_bidx = $1 LIMIT 1`
	err := ur.database.Database().QueryRowContext(ctx, query, email_bidx).Scan(
		&existingUser.ID, &existingUser.Email, &existingUser.Phone,
	)
	if err != nil {
		fmt.Println("Error fetching user by email:", err)
		return nil, err
	}

	return &existingUser, nil
}

func NewUserRepository(db pqsql.Client) domain.UserRepository {
	return &useRepository{
		database: db,
	}
}
