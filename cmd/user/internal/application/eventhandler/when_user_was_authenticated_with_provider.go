package eventhandler

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/mar1n3r0/go-api-boilerplate/cmd/user/internal/domain/user"
	"github.com/mar1n3r0/go-api-boilerplate/cmd/user/internal/infrastructure/persistence"
	"github.com/mar1n3r0/go-api-boilerplate/pkg/domain"
	"github.com/mar1n3r0/go-api-boilerplate/pkg/eventbus"
)

// WhenUserWasRegisteredWithProvider handles event
func WhenUserWasRegisteredWithProvider(db *sql.DB, repository persistence.UserRepository) eventbus.EventHandler {
	fn := func(ctx context.Context, event domain.Event) {
		// this goroutine runs independently to request's goroutine,
		// there for recover middlewears will not recover from panic to prevent crash
		defer recoverEventHandler()

		log.Printf("[EventHandler] %s\n", event.Payload)

		e := user.WasAuthenticatedWithProvider{}

		err := json.Unmarshal(event.Payload, &e)
		if err != nil {
			log.Printf("[EventHandler] Error: %v\n", err)
			return
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			log.Printf("[EventHandler] Error: %v\n", err)
			return
		}
		defer tx.Rollback()

		err = repository.Add(ctx, userWasAuthenticatedWithProviderModel{e})
		if err != nil {
			log.Printf("[EventHandler] Error: %v\n", err)
			return
		}

		tx.Commit()
	}

	return fn
}

type userWasAuthenticatedWithProviderModel struct {
	e user.WasAuthenticatedWithProvider
}

// GetID the id
func (u userWasAuthenticatedWithProviderModel) GetID() string {
	return u.e.ID.String()
}

// GetProvider the provider
func (u userWasAuthenticatedWithProviderModel) GetProvider() string {
	return u.e.Provider
}

// GetName the full name
func (u userWasAuthenticatedWithProviderModel) GetName() string {
	return u.e.Name
}

// GetEmail the email
func (u userWasAuthenticatedWithProviderModel) GetEmail() string {
	return u.e.Email
}

// GetNickName the nickname
func (u userWasAuthenticatedWithProviderModel) GetNickName() string {
	return u.e.NickName
}

// GetLocation the location
func (u userWasAuthenticatedWithProviderModel) GetLocation() string {
	return u.e.Location
}

// GetAvatarURL the avatarurl
func (u userWasAuthenticatedWithProviderModel) GetAvatarURL() string {
	return u.e.AvatarURL
}

// GetDescription the description
func (u userWasAuthenticatedWithProviderModel) GetDescription() string {
	return u.e.Description
}

// GetUserID the userid
func (u userWasAuthenticatedWithProviderModel) GetUserID() string {
	return u.e.UserID
}

// GetRefreshToken the refreshtoken
func (u userWasAuthenticatedWithProviderModel) GetRefreshToken() string {
	return u.e.RefreshToken
}
