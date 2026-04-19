package handlers

import (
	"context"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"kommande/internal/middleware"
	"kommande/internal/models"
)

func (h *Handler) ProfilePage(w http.ResponseWriter, r *http.Request) {
	jwtUser := middleware.GetUser(r.Context())

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var user models.User
	if err := h.db.Collection("users").FindOne(ctx, bson.M{"_id": jwtUser.ID}).Decode(&user); err != nil {
		http.Error(w, "Utilisateur introuvable", http.StatusNotFound)
		return
	}

	h.render(w, r, "templates/profile.html", "Mon profil", &user)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	jwtUser := middleware.GetUser(r.Context())

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	if email == "" {
		setFlash(w, "L'email est obligatoire.", "danger")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	update := bson.M{"email": email}

	photoID, err := h.uploadImage(ctx, r)
	if err == nil {
		update["photo_id"] = photoID
	}

	if _, err := h.db.Collection("users").UpdateOne(ctx,
		bson.M{"_id": jwtUser.ID},
		bson.M{"$set": update},
	); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Re-fetch and refresh JWT so nav shows updated photo immediately
	var updated models.User
	if err := h.db.Collection("users").FindOne(ctx, bson.M{"_id": jwtUser.ID}).Decode(&updated); err == nil {
		if token, err := h.generateJWT(&updated); err == nil {
			http.SetCookie(w, &http.Cookie{
				Name:     "auth_token",
				Value:    token,
				Path:     "/",
				MaxAge:   86400 * 7,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
		}
	}

	setFlash(w, "Profil mis à jour.", "success")
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
