package auth

import (
	"errors"
	"net/http"
	"strings"
)

var (
	ErrInvalidAPIKey = errors.New("invalid api key")
)

const (
	AuthHeader         = "Authorization"
	ApiKeyHeaderPrefix = "api-key"

	ScopeReadEvents  = "read:events"
	ScopeWriteEvents = "write:events"
)

func Authorize(validator ApiKeyValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Logic here
			apiKey := ""
			authH := r.Header[AuthHeader]

			if len(authH) > 0 {
				authHV := authH[0]
				authParts := strings.Split(authHV, " ")
				if len(authParts) == 2 {
					if strings.ToLower(authParts[0]) == ApiKeyHeaderPrefix {
						apiKey = authParts[1]
					}
				}
			}

			if apiKey == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			md, err := validator.Do(apiKey)
			if err != nil || md.Empty() {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			r = r.WithContext(withAuthMetadata(r.Context(), md))

			// Call the next handler
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

type ApiKeyValidator interface {
	Do(apiKey string) (Metadata, error)
}
