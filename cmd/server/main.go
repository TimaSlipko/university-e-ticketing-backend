// cmd/server/main.go
package main

import (
	"context"
	"errors"
	"eticketing/internal/models"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"eticketing/internal/config"
	"eticketing/internal/database"
	"eticketing/internal/handlers"
	"eticketing/internal/middleware"
	"eticketing/internal/repositories"
	"eticketing/internal/services"
	"eticketing/internal/utils"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.AutoMigrate(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize dependencies
	jwtManager := utils.NewJWTManager(&cfg.JWT)

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db.DB)
	sellerRepo := repositories.NewSellerRepository(db.DB)
	adminRepo := repositories.NewAdminRepository(db.DB)
	eventRepo := repositories.NewEventRepository(db.DB)
	ticketRepo := repositories.NewTicketRepository(db.DB)
	purchasedTicketRepo := repositories.NewPurchasedTicketRepository(db.DB)
	paymentRepo := repositories.NewPaymentRepository(db.DB)
	transferRepo := repositories.NewTransferRepository(db.DB)

	// Initialize services
	authService := services.NewAuthService(userRepo, sellerRepo, adminRepo, jwtManager)
	userService := services.NewUserService(userRepo)
	sellerService := services.NewSellerService(sellerRepo, eventRepo)
	adminService := services.NewAdminService(adminRepo, userRepo, sellerRepo, eventRepo, paymentRepo)
	paymentService := services.NewPaymentService(paymentRepo, cfg.Payment.IsMocked)
	eventService := services.NewEventService(eventRepo, ticketRepo)
	ticketService := services.NewTicketService(ticketRepo, purchasedTicketRepo, eventRepo, paymentService)
	transferService := services.NewTransferService(transferRepo, purchasedTicketRepo, userRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	sellerHandler := handlers.NewSellerHandler(sellerService)
	adminHandler := handlers.NewAdminHandler(adminService)
	eventHandler := handlers.NewEventHandler(eventService)
	ticketHandler := handlers.NewTicketHandler(ticketService)
	transferHandler := handlers.NewTransferHandler(transferService)

	gin.SetMode(gin.ReleaseMode)

	// Initialize router
	router := setupRouter(
		authHandler,
		userHandler,
		sellerHandler,
		adminHandler,
		eventHandler,
		ticketHandler,
		transferHandler,
		jwtManager,
	)

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Failed to start server:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown server gracefully
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		// Force shutdown if graceful shutdown fails
		if err := server.Close(); err != nil {
			log.Printf("Error closing server: %v", err)
		}
	} else {
		log.Println("Server gracefully stopped")
	}

	log.Println("Server exited")
}

func setupRouter(
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	sellerHandler *handlers.SellerHandler,
	adminHandler *handlers.AdminHandler,
	eventHandler *handlers.EventHandler,
	ticketHandler *handlers.TicketHandler,
	transferHandler *handlers.TransferHandler,
	jwtManager *utils.JWTManager,
) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.CORSMiddleware())

	// Rate limiting middleware
	router.Use(middleware.RateLimitMiddleware(time.Minute, 200))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
		}

		// Events routes (public for viewing)
		events := api.Group("/events")
		{
			events.GET("", eventHandler.GetEvents)
			events.GET("/:event_id", eventHandler.GetEvent)
			events.GET("/:event_id/tickets", ticketHandler.GetEventTickets)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(jwtManager))
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("/profile", userHandler.GetProfile)
				users.PUT("/profile", userHandler.UpdateProfile)
				users.PUT("/password", userHandler.ChangePassword)
				users.DELETE("/profile", userHandler.DeleteAccount)
			}

			// Ticket routes
			tickets := protected.Group("/tickets")
			{
				tickets.POST("/purchase", ticketHandler.PurchaseTicket)
				tickets.GET("/my", ticketHandler.GetMyTickets)
				tickets.POST("/transfer", transferHandler.InitiateTransfer) // Updated to use transferHandler
			}

			// Transfer routes
			transfers := protected.Group("/transfers")
			{
				transfers.GET("/active", transferHandler.GetActiveTransfers)
				transfers.POST("/:transfer_id/accept", transferHandler.AcceptTransfer)
				transfers.POST("/:transfer_id/reject", transferHandler.RejectTransfer)
				transfers.GET("/history", transferHandler.GetTransferHistory)
			}

			// Seller routes
			seller := protected.Group("/seller")
			seller.Use(middleware.RequireRole(models.UserTypeSeller))
			{
				seller.GET("/profile", sellerHandler.GetProfile)
				seller.PUT("/profile", sellerHandler.UpdateProfile)
				seller.PUT("/password", sellerHandler.ChangePassword)
				seller.DELETE("/profile", sellerHandler.DeleteAccount)

				seller.POST("/events", eventHandler.CreateEvent)
				seller.GET("/events", eventHandler.GetMyEvents)
				seller.PUT("/events/:event_id", eventHandler.UpdateEvent)
				seller.DELETE("/events/:event_id", eventHandler.DeleteEvent)
			}

			// Admin routes
			admin := protected.Group("/admin")
			admin.Use(middleware.RequireRole(models.UserTypeAdmin))
			{
				admin.GET("/events/pending", adminHandler.GetPendingEvents)
				admin.POST("/events/:event_id/approve", adminHandler.ApproveEvent)
				admin.POST("/events/:event_id/reject", adminHandler.RejectEvent)
				admin.GET("/stats", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Admin stats - not implemented yet"})
				})
			}
		}
	}

	return router
}
