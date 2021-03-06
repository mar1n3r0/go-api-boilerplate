/*
Package mysql holds view model repositories
*/
package mysql

import (
	"context"
	"database/sql"

	"github.com/mar1n3r0/go-api-boilerplate/cmd/user/internal/infrastructure/persistence"
	"github.com/mar1n3r0/go-api-boilerplate/pkg/errors"
	"github.com/mar1n3r0/go-api-boilerplate/pkg/mysql"
)

// NewUserRepository returns mysql view model repository for user
func NewUserRepository(db *sql.DB) persistence.UserRepository {
	return &userRepository{db}
}

type userRepository struct {
	db *sql.DB
}

func (r *userRepository) FindAll(ctx context.Context, limit, offset int32) ([]persistence.User, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, provider, name, emailAddress, nickName, location, avatarURL, description, userid, refreshToken FROM users ORDER BY id DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, errors.INTERNAL, "Could not query database")
	}
	defer rows.Close()

	var users []persistence.User

	for rows.Next() {
		user := User{}
		err = rows.Scan(&user.ID, &user.Provider, &user.Name, &user.Email, &user.NickName, &user.Location, &user.AvatarURL, &user.Description, &user.UserID, &user.RefreshToken)
		if err != nil {
			return nil, errors.Wrap(err, errors.INTERNAL, "Error while scanning users table")
		}

		users = append(users, user)
	}

	err = rows.Err()
	if err != nil {
		return nil, errors.New(errors.INTERNAL, "Error while getting rows")
	}

	return users, nil
}

func (r *userRepository) Get(ctx context.Context, id string) (persistence.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, provider, name, emailAddress, nickName, location, avatarURL, description, userId, refreshToken FROM users WHERE id=? LIMIT 1`, id)

	user := User{}

	err := row.Scan(&user.ID, &user.Provider, &user.Name, &user.Email, &user.NickName, &user.Location, &user.AvatarURL, &user.Description, &user.UserID, &user.RefreshToken)
	switch {
	case err == sql.ErrNoRows:
		return nil, errors.Wrap(err, errors.NOTFOUND, "User not found")
	case err != nil:
		return nil, errors.Wrap(err, errors.INTERNAL, "Error while scanning users table")
	default:
		return user, nil
	}
}

func (r *userRepository) Add(ctx context.Context, u persistence.User) error {
	user := User{
		ID:    u.GetID(),
		Email: u.GetEmail(),
		Provider: mysql.NullString{NullString: sql.NullString{
			String: u.GetProvider(),
			Valid:  u.GetProvider() != "",
		}},
		Name: mysql.NullString{NullString: sql.NullString{
			String: u.GetName(),
			Valid:  u.GetName() != "",
		}},
		NickName: mysql.NullString{NullString: sql.NullString{
			String: u.GetNickName(),
			Valid:  u.GetNickName() != "",
		}},
		Location: mysql.NullString{NullString: sql.NullString{
			String: u.GetLocation(),
			Valid:  u.GetLocation() != "",
		}},
		AvatarURL: mysql.NullString{NullString: sql.NullString{
			String: u.GetAvatarURL(),
			Valid:  u.GetAvatarURL() != "",
		}},
		Description: mysql.NullString{NullString: sql.NullString{
			String: u.GetDescription(),
			Valid:  u.GetDescription() != "",
		}},
		UserID: mysql.NullString{NullString: sql.NullString{
			String: u.GetUserID(),
			Valid:  u.GetUserID() != "",
		}},
		RefreshToken: mysql.NullString{NullString: sql.NullString{
			String: u.GetRefreshToken(),
			Valid:  u.GetRefreshToken() != "",
		}},
	}

	stmt, err := r.db.PrepareContext(ctx, `INSERT INTO users (id, provider, name, emailAddress, nickName, location, avatarURL, description, userID, refreshToken) VALUES (?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return errors.Wrap(err, errors.INTERNAL, "Invalid user insert query")
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, user.ID, user.Provider, user.Name, user.Email, user.NickName, user.Location, user.AvatarURL, user.Description, user.UserID, user.RefreshToken)
	if err != nil {
		return errors.Wrap(err, errors.INTERNAL, "Could not add user")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, errors.INTERNAL, "Could not get affected rows")
	}

	if rows != 1 {
		return errors.New(errors.INTERNAL, "Did not add user")
	}

	return nil
}

func (r *userRepository) UpdateEmail(ctx context.Context, id, email string) error {
	stmt, err := r.db.PrepareContext(ctx, `UPDATE users SET emailAddress=? WHERE id=?`)
	if err != nil {
		return errors.Wrap(err, errors.INTERNAL, "Invalid user update query")
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, email, id)
	if err != nil {
		return errors.Wrap(err, errors.INTERNAL, "Could not update user")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, errors.INTERNAL, "Could not get affected rows")
	}

	if rows != 1 {
		return errors.New(errors.INTERNAL, "Did not update user")
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	stmt, err := r.db.PrepareContext(ctx, `DELETE FROM users WHERE id=?`)
	if err != nil {
		return errors.Wrap(err, errors.INTERNAL, "Invalid user delete query")
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, id)
	if err != nil {
		return errors.Wrap(err, errors.INTERNAL, "Could not delete user")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, errors.INTERNAL, "Could not get affected rows")
	}

	if rows != 1 {
		return errors.New(errors.INTERNAL, "Did not delete user")
	}

	return nil
}

func (r *userRepository) Count(ctx context.Context) (int32, error) {
	var totalUsers int32

	row := r.db.QueryRowContext(ctx, `SELECT COUNT(distinctId) FROM users`)
	err := row.Scan(&totalUsers)
	if err != nil {
		return 0, errors.Wrap(err, errors.INTERNAL, "Could not count users")
	}

	return totalUsers, nil
}
