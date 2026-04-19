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
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		setFlash(w, "Le nom est obligatoire.", "danger")
		http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	cat := models.Category{
		ID:        bson.NewObjectID(),
		Name:      name,
		CreatedAt: time.Now(),
	}
	if imageID, err := h.uploadImage(ctx, r); err == nil {
		cat.ImageID = &imageID
	}

	if _, err := h.db.Collection("categories").InsertOne(ctx, cat); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	setFlash(w, "Catégorie créée.", "success")
	http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}

func (h *Handler) AdminEditCategory(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var cat models.Category
	if err := h.db.Collection("categories").FindOne(ctx, bson.M{"_id": id}).Decode(&cat); err != nil {
		http.NotFound(w, r)
		return
	}

	h.render(w, r, "templates/admin/category-form.html", "Modifier "+cat.Name, &cat)
}

func (h *Handler) AdminUpdateCategory(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		setFlash(w, "Le nom est obligatoire.", "danger")
		http.Redirect(w, r, "/admin/categories/"+idStr+"/edit", http.StatusSeeOther)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	update := bson.M{"name": name}
	if imageID, err := h.uploadImage(ctx, r); err == nil {
		update["image_id"] = imageID
	}

	if _, err := h.db.Collection("categories").UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": update},
	); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	setFlash(w, "Catégorie mise à jour.", "success")
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
