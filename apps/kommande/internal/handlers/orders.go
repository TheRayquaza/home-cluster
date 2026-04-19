package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"kommande/internal/email"
	"kommande/internal/middleware"
	"kommande/internal/models"
)

type orderPageData struct {
	Groups []models.ArticleGroup
	Order  *models.Order
	Date   string
	Today  string
}

func (h *Handler) OrderPage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	today := time.Now().Format("2006-01-02")

	date := r.URL.Query().Get("date")
	if date == "" {
		date = today
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		date = today
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

	var existingOrder *models.Order
	var order models.Order
	if err := h.db.Collection("orders").FindOne(ctx, bson.M{
		"username": user.Username,
		"date":     date,
	}).Decode(&order); err == nil {
		existingOrder = &order
	}

	h.render(w, r, "templates/order.html", "Commander", orderPageData{
		Groups: groupByCategory(h.loadCategories(ctx), articles),
		Order:  existingOrder,
		Date:   date,
		Today:  today,
	})
}

func (h *Handler) SubmitOrder(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	today := time.Now().Format("2006-01-02")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	date := r.FormValue("date")
	if date == "" {
		date = today
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		setFlash(w, "Date invalide.", "danger")
		http.Redirect(w, r, "/order", http.StatusSeeOther)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
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

	var items []models.OrderItem
	for _, a := range articles {
		qtyStr := r.FormValue(fmt.Sprintf("qty_%s", a.ID.Hex()))
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
			ImageID:     a.ImageID,
		})
	}

	if len(items) == 0 {
		setFlash(w, "Sélectionnez au moins un article.", "danger")
		http.Redirect(w, r, "/order?date="+date, http.StatusSeeOther)
		return
	}

	now := time.Now()

	var existing models.Order
	err = h.db.Collection("orders").FindOne(ctx, bson.M{
		"username": user.Username,
		"date":     date,
	}).Decode(&existing)

	isNew := err != nil

	if !isNew {
		if existing.Status != "pending" {
			setFlash(w, "Votre commande est déjà "+statusLabel(existing.Status)+". Contactez l'administrateur pour toute modification.", "warning")
			http.Redirect(w, r, "/order?date="+date, http.StatusSeeOther)
			return
		}
		_, err = h.db.Collection("orders").UpdateOne(ctx,
			bson.M{"_id": existing.ID},
			bson.M{"$set": bson.M{"items": items, "updated_at": now}},
		)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		setFlash(w, "Commande mise à jour !", "success")
	} else {
		if date < today {
			setFlash(w, "Vous ne pouvez pas commander pour une date passée.", "danger")
			http.Redirect(w, r, "/order", http.StatusSeeOther)
			return
		}
		newOrder := models.Order{
			ID:        bson.NewObjectID(),
			UserID:    user.ID,
			Username:  user.Username,
			Date:      date,
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

	h.notifyAdminNewOrder(user.Username, date, items, isNew)

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

func (h *Handler) notifyAdminNewOrder(username, date string, items []models.OrderItem, isNew bool) {
	action := "mise à jour"
	if isNew {
		action = "nouvelle"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Commande %s de %s pour le %s :\n\n", action, username, date)
	for _, item := range items {
		fmt.Fprintf(&sb, "  - %s : %s %s\n", item.ArticleName, formatQty(item.Quantity), item.Unit)
	}
	email.Send(h.cfg, h.cfg.AdminEmail, fmt.Sprintf("[Kommande] Commande de %s — %s", username, date), sb.String())
}
