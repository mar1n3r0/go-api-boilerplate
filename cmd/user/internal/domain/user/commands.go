package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"runtime/debug"

	"github.com/google/uuid"
	"github.com/thedevsaddam/govalidator"

	"github.com/mar1n3r0/go-api-boilerplate/pkg/commandbus"
	"github.com/mar1n3r0/go-api-boilerplate/pkg/domain"
	"github.com/mar1n3r0/go-api-boilerplate/pkg/errors"
	"github.com/mar1n3r0/go-api-boilerplate/pkg/executioncontext"
)

const (
	// RequestUserAccessToken command bus contract
	RequestUserAccessToken = "request-user-access-token"
	// ChangeUserEmailAddress command bus contract
	ChangeUserEmailAddress = "change-user-email-address"
	// RegisterUserWithEmail command bus contract
	RegisterUserWithEmail = "register-user-with-email"
	// RegisterUserWithProvider command bus contract
	RegisterUserWithProvider = "register-user-with-provider"
)

// NewCommandFromPayload builds command by contract from json payload
func NewCommandFromPayload(contract string, payload []byte) (domain.Command, error) {
	switch contract {
	case RegisterUserWithEmail:
		registerWithEmail := RegisterWithEmail{}
		err := unmarshalPayload(payload, &registerWithEmail)
		// validation rules
		rules := govalidator.MapData{
			"name":  []string{"required", "min:8", "max:32", "alpha_space"},
			"email": []string{"required", "min:8", "max:32", "email"},
		}

		opts := govalidator.Options{
			Data:  &registerWithEmail,
			Rules: rules,
		}

		v := govalidator.New(opts)
		e := v.ValidateStruct()
		if len(e) > 0 {
			data, _ := json.MarshalIndent(e, "", "  ")
			return nil, errors.New(errors.INVALID, string(data))
		}

		return registerWithEmail, err
	case RegisterUserWithProvider:
		registerWithProvider := RegisterWithProvider{}
		err := unmarshalPayload(payload, &registerWithProvider)

		return registerWithProvider, err
	case ChangeUserEmailAddress:
		changeEmailAddress := ChangeEmailAddress{}
		err := unmarshalPayload(payload, &changeEmailAddress)

		return changeEmailAddress, err
	case RequestUserAccessToken:
		requestAccessToken := RequestAccessToken{}
		err := unmarshalPayload(payload, &requestAccessToken)

		return requestAccessToken, err
	default:
		return nil, errors.New(errors.INTERNAL, "Invalid command contract")
	}
}

// RequestAccessToken command
type RequestAccessToken struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

// GetName returns command name
func (c RequestAccessToken) GetName() string {
	return fmt.Sprintf("%T", c)
}

// OnRequestAccessToken creates command handler
func OnRequestAccessToken(repository Repository, db *sql.DB) commandbus.CommandHandler {
	fn := func(ctx context.Context, c RequestAccessToken, out chan<- error) {
		// this goroutine runs independently to request's goroutine,
		// therefor recover middlewears will not recover from panic to prevent crash
		defer recoverCommandHandler(out)

		var id string

		row := db.QueryRowContext(ctx, `SELECT id FROM users WHERE emailAddress=?`, c.Email)
		err := row.Scan(&id)
		if err != nil {
			out <- errors.Wrap(err, errors.INTERNAL, "Could not ensure that user exists")
			return
		}

		u := repository.Get(uuid.MustParse(id))
		err = u.RequestAccessToken()
		if err != nil {
			out <- errors.Wrap(err, errors.INTERNAL, "Error when requesting access token")
			return
		}

		out <- repository.Save(executioncontext.WithFlag(context.Background(), executioncontext.LIVE), u)
	}

	return commandbus.CommandHandler(fn)
}

// ChangeEmailAddress command
type ChangeEmailAddress struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

// GetName returns command name
func (c ChangeEmailAddress) GetName() string {
	return fmt.Sprintf("%T", c)
}

// OnChangeEmailAddress creates command handler
func OnChangeEmailAddress(repository Repository, db *sql.DB) commandbus.CommandHandler {
	fn := func(ctx context.Context, c ChangeEmailAddress, out chan<- error) {
		// this goroutine runs independently to request's goroutine,
		// therefor recover middlewears will not recover from panic to prevent crash
		defer recoverCommandHandler(out)

		var totalUsers int32

		row := db.QueryRowContext(ctx, `SELECT COUNT(distinctId) FROM users WHERE emailAddress = ?`, c.Email)
		err := row.Scan(&totalUsers)
		if err != nil {
			out <- errors.Wrap(err, errors.INTERNAL, "Could not ensure that email is not taken")
			return
		}

		if totalUsers != 0 {
			out <- errors.Wrap(err, errors.INVALID, "User with given email already registered")
			return
		}

		u := repository.Get(c.ID)
		err = u.ChangeEmailAddress(c.Email)
		if err != nil {
			out <- errors.Wrap(err, errors.INTERNAL, "Error when changing email address")
			return
		}

		out <- repository.Save(executioncontext.WithFlag(context.Background(), executioncontext.LIVE), u)
	}

	return commandbus.CommandHandler(fn)
}

// RegisterWithEmail command
type RegisterWithEmail struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// GetName returns command name
func (c RegisterWithEmail) GetName() string {
	return fmt.Sprintf("%T", c)
}

// OnRegisterWithEmail creates command handler
func OnRegisterWithEmail(repository Repository, db *sql.DB) commandbus.CommandHandler {
	fn := func(ctx context.Context, c RegisterWithEmail, out chan<- error) {
		// this goroutine runs independently to request's goroutine,
		// therefor recover middlewears will not recover from panic to prevent crash
		defer recoverCommandHandler(out)

		var totalUsers int32

		row := db.QueryRowContext(ctx, `SELECT COUNT(distinctId) FROM users WHERE emailAddress = ?`, c.Email)
		err := row.Scan(&totalUsers)
		if err != nil {
			out <- errors.Wrap(err, errors.INTERNAL, "Could not ensure that email is not taken")
			return
		}

		if totalUsers != 0 {
			out <- errors.Wrap(err, errors.INVALID, "User with given email already registered")
			return
		}

		id, err := uuid.NewRandom()
		if err != nil {
			out <- errors.Wrap(err, errors.INTERNAL, "Could not generate new id")
			return
		}

		u := New()
		err = u.RegisterWithEmail(id, c.Name, c.Email)
		if err != nil {
			out <- errors.Wrap(err, errors.INTERNAL, "Error when registering new user")
			return
		}

		out <- repository.Save(executioncontext.WithFlag(context.Background(), executioncontext.LIVE), u)
	}

	return commandbus.CommandHandler(fn)
}

// RegisterWithProvider creates command handler
type RegisterWithProvider struct {
	ID           uuid.UUID `json:"id"`
	Provider     string    `json:"provider"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	NickName     string    `json:"nickName"`
	Location     string    `json:"location"`
	AvatarURL    string    `json:"avatarURL"`
	Description  string    `json:"description"`
	UserID       string    `json:"userId"`
	RefreshToken string    `json:"refreshToken"`
}

// GetName returns command name
func (c RegisterWithProvider) GetName() string {
	return fmt.Sprintf("%T", c)
}

// OnRegisterWithProvider creates command handler
func OnRegisterWithProvider(repository Repository, db *sql.DB) commandbus.CommandHandler {
	fn := func(ctx context.Context, c RegisterWithProvider, out chan<- error) {
		// this goroutine runs independently to request's goroutine,
		// therefor recover middlewears will not recover from panic to prevent crash
		defer recoverCommandHandler(out)

		var totalUsers int32

		row := db.QueryRowContext(ctx, `SELECT COUNT(distinctId) FROM users WHERE emailAddress = ?`, c.Email)
		err := row.Scan(&totalUsers)
		if err != nil {
			out <- errors.Wrap(err, errors.INTERNAL, "Could not ensure that user is not already registered")
			return
		}

		if totalUsers != 0 {
			out <- errors.Wrap(err, errors.INVALID, "User with given email already registered")
			return
		}

		u := New()
		err = u.RegisterWithProvider(c.ID, c.Provider, c.Name, c.Email, c.NickName, c.Location, c.AvatarURL, c.Description, c.UserID, c.RefreshToken)
		if err != nil {
			out <- errors.Wrap(err, errors.INTERNAL, "Error when registering new user")
			return
		}

		out <- repository.Save(executioncontext.WithFlag(context.Background(), executioncontext.LIVE), u)
	}

	return commandbus.CommandHandler(fn)
}

func recoverCommandHandler(out chan<- error) {
	if r := recover(); r != nil {
		out <- errors.Newf(errors.INTERNAL, "[CommandHandler] Recovered in %v", r)

		// Log the Go stack trace for this panic'd goroutine.
		log.Printf("%s\n", debug.Stack())
	}
}
