package eventstore

import "github.com/mar1n3r0/go-api-boilerplate/pkg/errors"

// ErrEventNotFound is thrown when an event is not found in the store.
var ErrEventNotFound = errors.New(errors.NOTFOUND, "Event not found")
