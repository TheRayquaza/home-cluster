package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"kommande/internal/middleware"
	"kommande/internal/models"
)

type orderPageData struct {
	Articles   []models.Article
	TodayOrder *models.Order
	Today      string
}

func (h *Handler) OrderPage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	today := time.Now().Format("2006-01-02")

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

	// Load today's order for this user if it exists
	var todayOrder *models.Order
	var order models.Order
	err = h.db.Collection("orders").FindOne(ctx, bson.M{
		"username": user.Username,
		"date":     today,
	}).Decode(&order)
	if err == nil {
		todayOrder = &order
	}

	h.render(w, r, "templates/order.html", "Commander", orderPageData{
		Articles:   articles,
		TodayOrder: todayOrder,
		Today:      today,
	})
}

func (h *Handler) SubmitOrder(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	today := time.Now().Format("2006-01-02")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Fetch all available articles to validate IDs
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

	articleMap := make(map[string]models.Article, len(articles))
	for _, a := range articles {
		articleMap[a.ID.Hex()] = a
	}

	var items []models.OrderItem
	for _, a := range articles {
		key := fmt.Sprintf("qty_%s", a.ID.Hex())
		qtyStr := r.FormValue(key)
		if qtyStr == "" {
			continue
		}
		qty, err := strconv.ParseFloat(qtyStr, 64)
		if err != nil || qty <= 0 {
			continue
		}
		items = append(items, models.OrderItem{
			ArticleID:   a.ID,
			ArticleName: a.Name,
			Quantity:    qty,
			Unit:        a.Unit,
		})
	}

	if len(items) == 0 {
		setFlash(w, "Sélectionnez au moins un article.", "danger")
		http.Redirect(w, r, "/order", http.StatusSeeOther)
		return
	}

	now := time.Now()

	// Check if order already exists today
	var existing models.Order
	err = h.db.Collection("orders").FindOne(ctx, bson.M{
		"username": user.Username,
		"date":     today,
	}).Decode(&existing)

	if err == nil {
		// Update only if pending
		if existing.Status != "pending" {
			setFlash(w, "Votre commande est déjà "+statusLabel(existing.Status)+". Contactez l'administrateur pour toute modification.", "warning")
			http.Redirect(w, r, "/order", http.StatusSeeOther)
			return
		}
		_, err = h.db.Collection("orders").UpdateOne(ctx,
			bson.M{"_id": existing.ID},
			bson.M{"$set": bson.M{
				"items":      items,
				"updated_at": now,
			}},
		)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		setFlash(w, "Commande mise à jour !", "success")
	} else {
		// Create new order
		newOrder := models.Order{
			ID:        bson.NewObjectID(),
			UserID:    user.ID,
			Username:  user.Username,
			Date:      today,
			Items:     items,
			Status:    "pending",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if _, err := h.db.Collection("orders").InsertOne(ctx, newOrder); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		setFlash(w, "Commande envoyée !", "success")
	}

	http.Redirect(w, r, "/orders", http.StatusSeeOther)
}

func (h *Handler) MyOrders(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}}).SetLimit(30)
	cursor, err := h.db.Collection("orders").Find(ctx, bson.M{"username": user.Username}, opts)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	h.render(w, r, "templates/my-orders.html", "Mes commandes", orders)
}
