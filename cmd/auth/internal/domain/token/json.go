/*
Package token holds token domain logic
*/
package token

import (
	"encoding/json"

	"github.com/mar1n3r0/go-api-boilerplate/pkg/errors"
)

func unmarshalPayload(payload []byte, model interface{}) error {
	err := json.Unmarshal(payload, model)
	if err != nil {
		return errors.Wrapf(err, errors.INTERNAL, "Error while trying to unmarshal payload %s", payload)
	}

	return nil
}
