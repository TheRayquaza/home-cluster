package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"kommande/internal/config"
	dbpkg "kommande/internal/db"
	"kommande/internal/handlers"
	"kommande/internal/middleware"
	"kommande/internal/models"
)

//go:embed templates static
var files embed.FS

func main() {
	cfg := config.Load()

	client, err := dbpkg.Connect(cfg.MongoURI)
	if err != nil {
		log.Fatalf("MongoDB connection failed: %v", err)
	}

	database := client.Database(cfg.DBName)
	ctx := context.Background()

	// Indexes
	database.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	database.Collection("orders").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "username", Value: 1}, {Key: "date", Value: -1}},
	})

	seedAdmin(ctx, database)

	h := handlers.New(database, files, cfg)

	mux := http.NewServeMux()

	// Static assets
	mux.Handle("GET /static/", http.FileServer(http.FS(files)))

	// Auth
	mux.HandleFunc("GET /login", h.LoginPage)
	mux.HandleFunc("POST /login", h.Login)
	mux.HandleFunc("GET /register", h.RegisterPage)
	mux.HandleFunc("POST /register", h.Register)
	mux.HandleFunc("POST /logout", h.Logout)

	// User routes (require auth)
	mux.Handle("GET /", middleware.RequireAuth(cfg.JWTSecret, http.HandlerFunc(h.Index)))
	mux.Handle("GET /order", middleware.RequireAuth(cfg.JWTSecret, http.HandlerFunc(h.OrderPage)))
	mux.Handle("POST /order", middleware.RequireAuth(cfg.JWTSecret, http.HandlerFunc(h.SubmitOrder)))
	mux.Handle("GET /orders", middleware.RequireAuth(cfg.JWTSecret, http.HandlerFunc(h.MyOrders)))
	mux.Handle("GET /profile", middleware.RequireAuth(cfg.JWTSecret, http.HandlerFunc(h.ProfilePage)))
	mux.Handle("POST /profile", middleware.RequireAuth(cfg.JWTSecret, http.HandlerFunc(h.UpdateProfile)))
	mux.Handle("GET /images/{id}", middleware.RequireAuth(cfg.JWTSecret, http.HandlerFunc(h.ServeImage)))

	// Admin routes (require admin role)
	mux.Handle("GET /admin", middleware.RequireAdmin(cfg.JWTSecret, http.HandlerFunc(h.AdminDashboard)))
	mux.Handle("GET /admin/articles", middleware.RequireAdmin(cfg.JWTSecret, http.HandlerFunc(h.AdminArticles)))
	mux.Handle("GET /admin/articles/new", middleware.RequireAdmin(cfg.JWTSecret, http.HandlerFunc(h.AdminNewArticle)))
	mux.Handle("POST /admin/articles", middleware.RequireAdmin(cfg.JWTSecret, http.HandlerFunc(h.AdminCreateArticle)))
	mux.Handle("GET /admin/articles/{id}/edit", middleware.RequireAdmin(cfg.JWTSecret, http.HandlerFunc(h.AdminEditArticle)))
	mux.Handle("POST /admin/articles/{id}", middleware.RequireAdmin(cfg.JWTSecret, http.HandlerFunc(h.AdminUpdateArticle)))
	mux.Handle("POST /admin/articles/{id}/delete", middleware.RequireAdmin(cfg.JWTSecret, http.HandlerFunc(h.AdminDeleteArticle)))
	mux.Handle("GET /admin/categories", middleware.RequireAdmin(cfg.JWTSecret, http.HandlerFunc(h.AdminCategories)))
	mux.Handle("POST /admin/categories", middleware.RequireAdmin(cfg.JWTSecret, http.HandlerFunc(h.AdminCreateCategory)))
	mux.Handle("GET /admin/categories/{id}/edit", middleware.RequireAdmin(cfg.JWTSecret, http.HandlerFunc(h.AdminEditCategory)))
	mux.Handle("POST /admin/categories/{id}", middleware.RequireAdmin(cfg.JWTSecret, http.HandlerFunc(h.AdminUpdateCategory)))
	mux.Handle("POST /admin/categories/{id}/delete", middleware.RequireAdmin(cfg.JWTSecret, http.HandlerFunc(h.AdminDeleteCategory)))
	mux.Handle("GET /admin/orders", middleware.RequireAdmin(cfg.JWTSecret, http.HandlerFunc(h.AdminOrders)))
	mux.Handle("POST /admin/orders/{id}/respond", middleware.RequireAdmin(cfg.JWTSecret, http.HandlerFunc(h.AdminRespondOrder)))

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Kommande listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-sigCtx.Done()
	log.Println("Shutting down...")
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(shutCtx)
	client.Disconnect(shutCtx)
}

func seedAdmin(ctx context.Context, db *mongo.Database) {
	count, _ := db.Collection("users").CountDocuments(ctx, bson.M{})
	if count > 0 {
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("seed admin error: %v", err)
		return
	}
	_, err = db.Collection("users").InsertOne(ctx, models.User{
		ID:           bson.NewObjectID(),
		Username:     "admin",
		Email:        "admin@kommande.local",
		PasswordHash: string(hash),
		Role:         "admin",
		CreatedAt:    time.Now(),
	})
	if err != nil {
		log.Printf("seed admin insert error: %v", err)
		return
	}
	log.Println("Default admin created — username: admin / password: admin123")
}
