package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"kommande/internal/middleware"
	"kommande/internal/models"
)

func (h *Handler) OIDCLoginRedirect(w http.ResponseWriter, r *http.Request) {
	if u := middleware.GetUser(r.Context()); u != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	state, err := generateState()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "oidc_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, h.oauth2Config.AuthCodeURL(state), http.StatusFound)
}

func (h *Handler) OIDCCallback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie("oidc_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "oidc_state", Path: "/", MaxAge: -1})

	code := r.URL.Query().Get("code")
	oauth2Token, err := h.oauth2Config.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("OIDC code exchange failed: %v", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No ID token in response", http.StatusInternalServerError)
		return
	}
	idToken, err := h.oidcVerifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		log.Printf("OIDC ID token verification failed: %v", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	var claims struct {
		Email             string   `json:"email"`
		PreferredUsername string   `json:"preferred_username"`
		Groups            []string `json:"groups"`
	}
	if err := idToken.Claims(&claims); err != nil {
		log.Printf("OIDC claims extraction failed: %v", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	role := "user"
	for _, g := range claims.Groups {
		if g == "kommande-admins" {
			role = "admin"
			break
		}
	}

	user, err := h.upsertOIDCUser(r.Context(), claims.Email, claims.PreferredUsername, role)
	if err != nil {
		log.Printf("OIDC user upsert failed: %v", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	token, err := h.generateJWT(user)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "auth_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	logoutURL := h.cfg.OIDCIssuer + "/protocol/openid-connect/logout" +
		"?post_logout_redirect_uri=" + url.QueryEscape(h.cfg.BaseURL) +
		"&client_id=" + url.QueryEscape(h.cfg.OIDCClientID)
	http.Redirect(w, r, logoutURL, http.StatusSeeOther)
}

func (h *Handler) upsertOIDCUser(ctx context.Context, email, username, role string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if username == "" {
		username = email
	}

	filter := bson.M{"email": email}
	update := bson.M{
		"$set": bson.M{
			"username":   username,
			"role":       role,
		},
		"$setOnInsert": bson.M{
			"_id":        bson.NewObjectID(),
			"email":      email,
			"created_at": time.Now(),
		},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var user models.User
	err := h.db.Collection("users").FindOneAndUpdate(ctx, filter, update, opts).Decode(&user)
	return &user, err
}

func (h *Handler) generateJWT(user *models.User) (string, error) {
	claims := middleware.Claims{
		UserID:   user.ID.Hex(),
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	if user.PhotoID != nil {
		claims.PhotoID = user.PhotoID.Hex()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		log.Printf("JWT sign error: %v", err)
		return "", err
	}
	return signed, nil
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
