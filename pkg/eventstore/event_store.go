package eventstore

import (
	"github.com/google/uuid"

	"github.com/mar1n3r0/go-api-boilerplate/pkg/domain"
)

// EventStore methods allow to save, load events and event streams
type EventStore interface {
	Store([]domain.Event) error
	Get(uuid.UUID) (domain.Event, error)
	FindAll() []domain.Event
	GetStream(uuid.UUID, string) []domain.Event
}
