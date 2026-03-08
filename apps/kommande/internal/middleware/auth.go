package middleware

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"kommande/internal/models"
)

type contextKey string

const UserContextKey contextKey = "user"

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func RequireAuth(jwtSecret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := extractUser(r, jwtSecret)
		if err != nil {
			clearAuthCookie(w)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireAdmin(jwtSecret string, next http.Handler) http.Handler {
	return RequireAuth(jwtSecret, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r.Context())
		if user == nil || user.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}))
}

func GetUser(ctx context.Context) *models.User {
	u, _ := ctx.Value(UserContextKey).(*models.User)
	return u
}

func extractUser(r *http.Request, jwtSecret string) (*models.User, error) {
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		return nil, err
	}

	claims := &Claims{}
	_, err = jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	id, _ := bson.ObjectIDFromHex(claims.UserID)
	return &models.User{
		ID:       id,
		Username: claims.Username,
		Role:     claims.Role,
	}, nil
}

func clearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "auth_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}
