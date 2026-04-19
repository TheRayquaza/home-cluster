package handlers

import (
	"context"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"kommande/internal/models"
)

func (h *Handler) AdminCategories(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})
	cursor, err := h.db.Collection("categories").Find(ctx, bson.M{}, opts)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var cats []models.Category
	if err := cursor.All(ctx, &cats); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	h.render(w, r, "templates/admin/categories.html", "Catégories", cats)
}

func (h *Handler) AdminCreateCategory(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		setFlash(w, "Le nom est obligatoire.", "danger")
		http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err := h.db.Collection("categories").InsertOne(ctx, models.Category{
		ID:        bson.NewObjectID(),
		Name:      name,
		CreatedAt: time.Now(),
	})
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	setFlash(w, "Catégorie créée.", "success")
	http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}

func (h *Handler) AdminDeleteCategory(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	h.db.Collection("categories").DeleteOne(ctx, bson.M{"_id": id})
	// Remove category reference from all articles
	h.db.Collection("articles").UpdateMany(ctx,
		bson.M{"category_ids": id},
		bson.M{"$pull": bson.M{"category_ids": id}},
	)

	setFlash(w, "Catégorie supprimée.", "success")
	http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}
