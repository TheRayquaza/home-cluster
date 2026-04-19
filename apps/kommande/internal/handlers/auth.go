package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"

	"kommande/internal/middleware"
	"kommande/internal/models"
)

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	// Already logged in → redirect
	if u := middleware.GetUser(r.Context()); u != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	h.render(w, r, "templates/login.html", "Connexion", nil)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		setFlash(w, "Veuillez remplir tous les champs.", "danger")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var user models.User
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err := h.db.Collection("users").FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		setFlash(w, "Identifiants incorrects.", "danger")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		setFlash(w, "Identifiants incorrects.", "danger")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	token, err := h.generateJWT(&user)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "templates/register.html", "Inscription", nil)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if username == "" || email == "" || password == "" {
		setFlash(w, "Veuillez remplir tous les champs.", "danger")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	if len(password) < 6 {
		setFlash(w, "Le mot de passe doit faire au moins 6 caractères.", "danger")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Check for existing username
	count, _ := h.db.Collection("users").CountDocuments(ctx, bson.M{"username": username})
	if count > 0 {
		setFlash(w, "Ce nom d'utilisateur est déjà pris.", "danger")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user := models.User{
		ID:           bson.NewObjectID(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         "user",
		CreatedAt:    time.Now(),
	}

	if _, err := h.db.Collection("users").InsertOne(ctx, user); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	token, err := h.generateJWT(&user)
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

	setFlash(w, "Bienvenue, "+username+" !", "success")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "auth_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
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
