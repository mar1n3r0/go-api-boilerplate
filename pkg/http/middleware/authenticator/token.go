package authenticator

import (
	"net/http"
	"strings"

	"github.com/mar1n3r0/go-api-boilerplate/pkg/errors"
	"github.com/mar1n3r0/go-api-boilerplate/pkg/http/response"
	"github.com/mar1n3r0/go-api-boilerplate/pkg/identity"
)

// TokenAuthFunc returns Identity from token
type TokenAuthFunc func(apiKey, token string) (identity.Identity, error)

// TokenAuthenticator authorize by token
// and adds Identity to request's Context
type TokenAuthenticator interface {
	// FromHeader authorize by the token provided in the request's Authorization header
	FromHeader(realm string) func(next http.Handler) http.Handler
	// FromQuery authorize by the token provided in the request's query parameter
	FromQuery(name string) func(next http.Handler) http.Handler
	// FromCookie authorize by the token provided in the request's cookie
	FromCookie(name string) func(next http.Handler) http.Handler
}

const (
	appKey = "secret"
)

type tokenAuth struct {
	afn TokenAuthFunc
}

func (a *tokenAuth) FromHeader(realm string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-Api-Key")
			if apiKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			if apiKey != appKey {
				response.RespondJSONError(r.Context(), w, errors.New(errors.UNAUTHORIZED, http.StatusText(http.StatusUnauthorized)))
				return
			}

			token := r.Header.Get("Authorization")
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			if strings.HasPrefix(token, "Bearer ") {
				//if bearer, err := base64.StdEncoding.DecodeString(token[7:]); err == nil {
				i, err := a.afn(apiKey, string(token[7:]))
				if err != nil {
					w.Header().Set("WWW-Authenticate", `Bearer realm="`+realm+`"`)
					response.RespondJSONError(r.Context(), w, errors.New(errors.UNAUTHORIZED, http.StatusText(http.StatusUnauthorized)))
					return
				}

				next.ServeHTTP(w, r.WithContext(identity.ContextWithIdentity(r.Context(), i)))
				return
				//}
			}

			w.Header().Set("WWW-Authenticate", `Bearer realm="`+realm+`"`)
			response.RespondJSONError(r.Context(), w, errors.New(errors.UNAUTHORIZED, http.StatusText(http.StatusUnauthorized)))
		}

		return http.HandlerFunc(fn)
	}
}

func (a *tokenAuth) FromQuery(name string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.URL.Query().Get("apiKey")
			if apiKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			if apiKey != appKey {
				response.RespondJSONError(r.Context(), w, errors.New(errors.UNAUTHORIZED, http.StatusText(http.StatusUnauthorized)))
				return
			}

			token := r.URL.Query().Get(name)
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			i, err := a.afn(apiKey, token)
			if err != nil {
				response.RespondJSONError(r.Context(), w, errors.New(errors.UNAUTHORIZED, http.StatusText(http.StatusUnauthorized)))
				return
			}

			next.ServeHTTP(w, r.WithContext(identity.ContextWithIdentity(r.Context(), i)))
		}

		return http.HandlerFunc(fn)
	}
}

func (a *tokenAuth) FromCookie(name string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(name)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			cookieValues := strings.Split(cookie.Value, "&")

			apiKey := cookieValues[0]
			token := cookieValues[1]

			i, err := a.afn(apiKey, token)
			if err != nil {
				response.RespondJSONError(r.Context(), w, errors.New(errors.UNAUTHORIZED, errors.ErrorMessage(err)))
				return
			}

			next.ServeHTTP(w, r.WithContext(identity.ContextWithIdentity(r.Context(), i)))
		}

		return http.HandlerFunc(fn)
	}
}

// NewToken returns new token authenticator
func NewToken(afn TokenAuthFunc) TokenAuthenticator {
	return &tokenAuth{
		afn: afn,
	}
}
