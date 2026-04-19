package handlers

import (
	"context"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"kommande/internal/models"
)

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	cursor, err := h.db.Collection("articles").Find(ctx, bson.M{"available": true})
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var articles []models.Article
	if err := cursor.All(ctx, &articles); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	groups := groupByCategory(h.loadCategories(ctx), articles)
	h.render(w, r, "templates/index.html", "Catalogue", groups)
}
