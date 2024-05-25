package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"unicode"

	"hestia/pkg/auth"
)

type ctxKey string

const authUser ctxKey = "authUserID"

type Middleware func(http.Handler) http.Handler

// CheckJWT is function to verify JWT token
func CheckJWT(i *auth.Interceptor, handler func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Missing authorization header")
			return
		}
		tokenString = tokenString[len("Bearer "):]

		method := combineURL(r.RequestURI)
		userID, err := i.Authorize(method, tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Invalid token")
			return
		}
		ctx := ContextWithUserID(r.Context(), userID)

		handler(w, r.WithContext(ctx))
	})
}

func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, authUser, userID)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(authUser).(string)
	if !ok {
		return "", false
	}

	return userID, true
}

func combineURL(url string) string {
	split := strings.Split(url, string(os.PathSeparator))
	lastEl := split[len(split)-1]
	if isInt(lastEl) {
		split = split[:len(split)-1]
		return strings.Join(split, string(os.PathSeparator))
	}
	return strings.Join(split, string(os.PathSeparator))
}

func isInt(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}
