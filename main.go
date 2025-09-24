package main

import (
	"log"
	"os"

	"ticket-booking-system/internal/config"
	"ticket-booking-system/internal/database"
	"ticket-booking-system/internal/handlers"
	"ticket-booking-system/internal/repository"
	"ticket-booking-system/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})

	// Initialize repositories
	eventRepo := repository.NewEventRepository(db)
	userRepo := repository.NewUserRepository(db)
	bookingRepo := repository.NewBookingRepository(db)

	// Initialize services
	eventService := services.NewEventService(eventRepo)
	userService := services.NewUserService(userRepo)
	bookingService := services.NewBookingService(bookingRepo, eventRepo, cfg.PaymentDeadline)
	paymentService := services.NewPaymentService(rdb, bookingRepo)

	// Start payment processor in background
	go paymentService.StartProcessor()

	// Start expired booking processor in background
	go paymentService.StartExpiredBookingProcessor()

	// Initialize handlers
	eventHandler := handlers.NewEventHandler(eventService)
	userHandler := handlers.NewUserHandler(userService)
	bookingHandler := handlers.NewBookingHandler(bookingService, paymentService)

	// Setup routes
	router := setupRoutes(eventHandler, userHandler, bookingHandler)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func setupRoutes(eventHandler *handlers.EventHandler, userHandler *handlers.UserHandler, bookingHandler *handlers.BookingHandler) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Event routes
		events := api.Group("/events")
		{
			events.GET("", eventHandler.GetEvents)
			events.GET("/:id", eventHandler.GetEvent)
			events.POST("", eventHandler.CreateEvent)
			events.PUT("/:id", eventHandler.UpdateEvent)
			events.DELETE("/:id", eventHandler.DeleteEvent)
			events.GET("/:id/statistics", eventHandler.GetEventStatistics)
		}

		// User routes
		users := api.Group("/users")
		{
			users.GET("", userHandler.GetUsers)
			users.GET("/:id", userHandler.GetUser)
			users.POST("", userHandler.CreateUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}

		// Booking routes
		bookings := api.Group("/bookings")
		{
			bookings.POST("", bookingHandler.CreateBooking)
			bookings.GET("/:id", bookingHandler.GetBooking)
			bookings.PUT("/:id/cancel", bookingHandler.CancelBooking)
			bookings.GET("/user/:user_id", bookingHandler.GetUserBookings)
		}
	}

	return router
}
