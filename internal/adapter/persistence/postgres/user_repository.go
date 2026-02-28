package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

// UserRepository implements repository.UserRepository using PostgreSQL.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, user.ID, user.Email.String(), user.PasswordHash.String(), user.Role, user.Status, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}

	_, err = db.Exec(ctx, `
		INSERT INTO user_profiles (user_id, display_name, avatar_url, date_of_birth, language, country, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, user.ID, user.Profile.DisplayName, user.Profile.AvatarURL, user.Profile.DateOfBirth,
		user.Profile.Language, user.Profile.Country, user.Profile.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating user profile: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	db := connFromCtx(ctx, r.pool)
	user := &entity.User{}
	var email, passwordHash string

	err := db.QueryRow(ctx, `
		SELECT u.id, u.email, u.password_hash, u.role, u.status, u.created_at, u.updated_at,
		       p.display_name, p.avatar_url, p.date_of_birth, p.language, p.country, p.updated_at
		FROM users u
		LEFT JOIN user_profiles p ON u.id = p.user_id
		WHERE u.id = $1
	`, id).Scan(
		&user.ID, &email, &passwordHash, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
		&user.Profile.DisplayName, &user.Profile.AvatarURL, &user.Profile.DateOfBirth,
		&user.Profile.Language, &user.Profile.Country, &user.Profile.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting user by id: %w", err)
	}

	user.Email = valueobject.EmailFromTrusted(email)
	user.PasswordHash = valueobject.NewHashedPassword(passwordHash)
	user.Profile.UserID = user.ID
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email valueobject.Email) (*entity.User, error) {
	db := connFromCtx(ctx, r.pool)
	user := &entity.User{}
	var emailStr, passwordHash string

	err := db.QueryRow(ctx, `
		SELECT u.id, u.email, u.password_hash, u.role, u.status, u.created_at, u.updated_at,
		       p.display_name, p.avatar_url, p.date_of_birth, p.language, p.country, p.updated_at
		FROM users u
		LEFT JOIN user_profiles p ON u.id = p.user_id
		WHERE u.email = $1
	`, email.String()).Scan(
		&user.ID, &emailStr, &passwordHash, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
		&user.Profile.DisplayName, &user.Profile.AvatarURL, &user.Profile.DateOfBirth,
		&user.Profile.Language, &user.Profile.Country, &user.Profile.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting user by email: %w", err)
	}

	user.Email = valueobject.EmailFromTrusted(emailStr)
	user.PasswordHash = valueobject.NewHashedPassword(passwordHash)
	user.Profile.UserID = user.ID
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		UPDATE users SET email = $2, password_hash = $3, role = $4, status = $5, updated_at = $6
		WHERE id = $1
	`, user.ID, user.Email.String(), user.PasswordHash.String(), user.Role, user.Status, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating user: %w", err)
	}

	_, err = db.Exec(ctx, `
		UPDATE user_profiles
		SET display_name = $2, avatar_url = $3, date_of_birth = $4, language = $5, country = $6, updated_at = $7
		WHERE user_id = $1
	`, user.ID, user.Profile.DisplayName, user.Profile.AvatarURL, user.Profile.DateOfBirth,
		user.Profile.Language, user.Profile.Country, user.Profile.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating user profile: %w", err)
	}

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `UPDATE users SET status = 'deleted', updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	return nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email valueobject.Email) (bool, error) {
	db := connFromCtx(ctx, r.pool)
	var exists bool
	err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email.String()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking email existence: %w", err)
	}
	return exists, nil
}

func (r *UserRepository) List(ctx context.Context, page, perPage int) ([]entity.User, int64, error) {
	db := connFromCtx(ctx, r.pool)
	offset := (page - 1) * perPage

	var total int64
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE status != 'deleted'`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("counting users: %w", err)
	}

	rows, err := db.Query(ctx, `
		SELECT u.id, u.email, u.role, u.status, u.created_at, u.updated_at,
		       p.display_name, p.avatar_url, p.language, p.country
		FROM users u
		LEFT JOIN user_profiles p ON u.id = p.user_id
		WHERE u.status != 'deleted'
		ORDER BY u.created_at DESC
		LIMIT $1 OFFSET $2
	`, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("listing users: %w", err)
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var u entity.User
		var email string
		err := rows.Scan(
			&u.ID, &email, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt,
			&u.Profile.DisplayName, &u.Profile.AvatarURL, &u.Profile.Language, &u.Profile.Country,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scanning user: %w", err)
		}
		u.Email = valueobject.EmailFromTrusted(email)
		u.Profile.UserID = u.ID
		users = append(users, u)
	}

	return users, total, nil
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.UserStatus) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `UPDATE users SET status = $2, updated_at = NOW() WHERE id = $1`, id, status)
	if err != nil {
		return fmt.Errorf("updating user status: %w", err)
	}
	return nil
}
