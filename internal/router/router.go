package router

import (
	"agenda-api/internal/config"
	"agenda-api/internal/handlers"
	"agenda-api/internal/middleware"
	"agenda-api/internal/models"
	"agenda-api/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func Setup(db *sqlx.DB, cfg *config.Config) *gin.Engine {
	gin.SetMode(cfg.GinMode)
	r := gin.Default()

	r.Use(middleware.CORS())

	userRepo := repository.NewUserRepository(db)
	eventRepo := repository.NewEventRepository(db)
	attendanceRepo := repository.NewAttendanceRepository(db)

	authHandler := handlers.NewAuthHandler(userRepo, cfg.JWTSecret, cfg.JWTExpirationHours)
	eventHandler := handlers.NewEventHandler(eventRepo)
	attendanceHandler := handlers.NewAttendanceHandler(attendanceRepo, eventRepo)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.JWTAuth(cfg.JWTSecret), authHandler.Me)
		}

		events := api.Group("/events")
		{
			events.GET("", eventHandler.GetAll)
			events.GET("/calendar", eventHandler.GetCalendar)
			events.GET("/:id", eventHandler.GetByID)

			events.POST("",
				middleware.JWTAuth(cfg.JWTSecret),
				middleware.RequireAdmin(),
				eventHandler.Create,
			)

			events.PATCH("/:id",
				middleware.JWTAuth(cfg.JWTSecret),
				middleware.RequireAdminOrSpeaker(),
				eventHandler.Update,
			)

			events.DELETE("/:id",
				middleware.JWTAuth(cfg.JWTSecret),
				middleware.RequireAdmin(),
				eventHandler.Delete,
			)

			events.POST("/:id/register",
				middleware.JWTAuth(cfg.JWTSecret),
				middleware.RequireRoles(models.RoleAttendee),
				attendanceHandler.Register,
			)

			events.DELETE("/:id/register",
				middleware.JWTAuth(cfg.JWTSecret),
				middleware.RequireRoles(models.RoleAttendee),
				attendanceHandler.Cancel,
			)

			events.GET("/:id/attendees",
				middleware.JWTAuth(cfg.JWTSecret),
				middleware.RequireAdminOrSpeaker(),
				attendanceHandler.GetAttendees,
			)
		}

		api.GET("/my-registrations",
			middleware.JWTAuth(cfg.JWTSecret),
			attendanceHandler.GetMyRegistrations,
		)
	}

	return r
}
