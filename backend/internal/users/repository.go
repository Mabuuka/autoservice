package users

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

type Repository struct {
	DB *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) CreateClient(ctx context.Context, input RegisterInput, passwordHash string) (AuthUser, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return AuthUser{}, err
	}
	defer tx.Rollback(ctx)

	var ownerID int64
	ownerQuery := `
		INSERT INTO automaster.owners (full_name, phone, driver_license, is_regular)
		VALUES ($1, $2, $3, false)
		RETURNING owner_id
	`

	if err := tx.QueryRow(ctx, ownerQuery, input.FullName, input.Phone, input.DriverLicense).Scan(&ownerID); err != nil {
		return AuthUser{}, err
	}

	userQuery := `
		INSERT INTO automaster.users (
			role,
			owner_id,
			employee_id,
			email,
			password_hash,
			full_name,
			phone,
			preferred_entrypoint
		)
		VALUES ('client', $1, NULL, $2, $3, $4, $5, 'profile')
		RETURNING
			user_id,
			role,
			owner_id,
			employee_id,
			email,
			password_hash,
			full_name,
			phone,
			preferred_entrypoint,
			TO_CHAR(created_at, 'YYYY-MM-DD"T"HH24:MI:SS') AS created_at,
			TO_CHAR(updated_at, 'YYYY-MM-DD"T"HH24:MI:SS') AS updated_at,
			COALESCE(TO_CHAR(last_login_at, 'YYYY-MM-DD"T"HH24:MI:SS'), '') AS last_login_at
	`

	user, err := scanAuthUserRow(tx.QueryRow(ctx, userQuery, ownerID, input.Email, passwordHash, input.FullName, input.Phone))
	if err != nil {
		return AuthUser{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return AuthUser{}, err
	}

	return user, nil
}

func (r *Repository) GetAuthUserByEmail(ctx context.Context, email string) (AuthUser, error) {
	query := `
		SELECT
			user_id,
			role,
			owner_id,
			employee_id,
			email,
			password_hash,
			full_name,
			phone,
			preferred_entrypoint,
			TO_CHAR(created_at, 'YYYY-MM-DD"T"HH24:MI:SS') AS created_at,
			TO_CHAR(updated_at, 'YYYY-MM-DD"T"HH24:MI:SS') AS updated_at,
			COALESCE(TO_CHAR(last_login_at, 'YYYY-MM-DD"T"HH24:MI:SS'), '') AS last_login_at
		FROM automaster.users
		WHERE email = $1
	`

	user, err := scanAuthUserRow(r.DB.QueryRow(ctx, query, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AuthUser{}, ErrUserNotFound
		}
		return AuthUser{}, err
	}

	return user, nil
}

func (r *Repository) GetAuthUserByID(ctx context.Context, userID int64) (AuthUser, error) {
	query := `
		SELECT
			user_id,
			role,
			owner_id,
			employee_id,
			email,
			password_hash,
			full_name,
			phone,
			preferred_entrypoint,
			TO_CHAR(created_at, 'YYYY-MM-DD"T"HH24:MI:SS') AS created_at,
			TO_CHAR(updated_at, 'YYYY-MM-DD"T"HH24:MI:SS') AS updated_at,
			COALESCE(TO_CHAR(last_login_at, 'YYYY-MM-DD"T"HH24:MI:SS'), '') AS last_login_at
		FROM automaster.users
		WHERE user_id = $1
	`

	user, err := scanAuthUserRow(r.DB.QueryRow(ctx, query, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AuthUser{}, ErrUserNotFound
		}
		return AuthUser{}, err
	}

	return user, nil
}

func (r *Repository) UpdateLastLogin(ctx context.Context, userID int64) error {
	query := `
		UPDATE automaster.users
		SET
			last_login_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1
	`

	result, err := r.DB.Exec(ctx, query, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *Repository) GetProfileByUserID(ctx context.Context, userID int64) (Profile, error) {
	authUser, err := r.GetAuthUserByID(ctx, userID)
	if err != nil {
		return Profile{}, err
	}

	profile := Profile{
		User: buildUser(authUser),
	}

	if authUser.Role == "client" && authUser.OwnerID != nil {
		owner, err := r.getOwnerProfile(ctx, *authUser.OwnerID)
		if err != nil {
			return Profile{}, err
		}
		profile.Owner = &owner
	}

	if authUser.Role == "master" && authUser.EmployeeID != nil {
		employee, err := r.getEmployeeProfile(ctx, *authUser.EmployeeID)
		if err != nil {
			return Profile{}, err
		}
		profile.Employee = &employee
	}

	return profile, nil
}

func (r *Repository) UpdateProfile(ctx context.Context, userID int64, input UpdateProfileInput) (Profile, error) {
	authUser, err := r.GetAuthUserByID(ctx, userID)
	if err != nil {
		return Profile{}, err
	}

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return Profile{}, err
	}
	defer tx.Rollback(ctx)

	updateUserQuery := `
		UPDATE automaster.users
		SET
			full_name = $2,
			phone = $3,
			preferred_entrypoint = $4,
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1
	`

	if _, err := tx.Exec(ctx, updateUserQuery, userID, input.FullName, input.Phone, input.PreferredEntrypoint); err != nil {
		return Profile{}, err
	}

	if authUser.Role == "client" && authUser.OwnerID != nil {
		updateOwnerQuery := `
			UPDATE automaster.owners
			SET
				full_name = $2,
				phone = $3,
				driver_license = $4
			WHERE owner_id = $1
		`

		if _, err := tx.Exec(ctx, updateOwnerQuery, *authUser.OwnerID, input.FullName, input.Phone, input.DriverLicense); err != nil {
			return Profile{}, err
		}
	}

	if authUser.Role == "master" && authUser.EmployeeID != nil {
		updateEmployeeQuery := `
			UPDATE automaster.employees
			SET
				full_name = $2,
				phone = $3
			WHERE employee_id = $1
		`

		if _, err := tx.Exec(ctx, updateEmployeeQuery, *authUser.EmployeeID, input.FullName, input.Phone); err != nil {
			return Profile{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Profile{}, err
	}

	return r.GetProfileByUserID(ctx, userID)
}

func (r *Repository) getOwnerProfile(ctx context.Context, ownerID int64) (OwnerProfile, error) {
	query := `
		SELECT owner_id, full_name, phone, driver_license, is_regular
		FROM automaster.owners
		WHERE owner_id = $1
	`

	var owner OwnerProfile
	if err := r.DB.QueryRow(ctx, query, ownerID).Scan(
		&owner.OwnerID,
		&owner.FullName,
		&owner.Phone,
		&owner.DriverLicense,
		&owner.IsRegular,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OwnerProfile{}, ErrUserNotFound
		}
		return OwnerProfile{}, err
	}

	return owner, nil
}

func (r *Repository) getEmployeeProfile(ctx context.Context, employeeID int64) (EmployeeProfile, error) {
	query := `
		SELECT employee_id, COALESCE(personnel_number, 0)::int, specialty, phone, full_name
		FROM automaster.employees
		WHERE employee_id = $1
	`

	var employee EmployeeProfile
	if err := r.DB.QueryRow(ctx, query, employeeID).Scan(
		&employee.EmployeeID,
		&employee.PersonnelNumber,
		&employee.Specialty,
		&employee.Phone,
		&employee.FullName,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EmployeeProfile{}, ErrUserNotFound
		}
		return EmployeeProfile{}, err
	}

	return employee, nil
}

func buildUser(authUser AuthUser) User {
	return User{
		UserID:              authUser.UserID,
		Role:                authUser.Role,
		OwnerID:             authUser.OwnerID,
		EmployeeID:          authUser.EmployeeID,
		Email:               authUser.Email,
		FullName:            authUser.FullName,
		Phone:               authUser.Phone,
		PreferredEntrypoint: authUser.PreferredEntrypoint,
		CreatedAt:           authUser.CreatedAt,
		UpdatedAt:           authUser.UpdatedAt,
		LastLoginAt:         authUser.LastLoginAt,
	}
}

type scanner interface {
	Scan(dest ...any) error
}

func scanAuthUserRow(row scanner) (AuthUser, error) {
	var user AuthUser
	var ownerID sql.NullInt64
	var employeeID sql.NullInt64

	err := row.Scan(
		&user.UserID,
		&user.Role,
		&ownerID,
		&employeeID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Phone,
		&user.PreferredEntrypoint,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)
	if err != nil {
		return AuthUser{}, err
	}

	if ownerID.Valid {
		value := ownerID.Int64
		user.OwnerID = &value
	}

	if employeeID.Valid {
		value := employeeID.Int64
		user.EmployeeID = &value
	}

	return user, nil
}
