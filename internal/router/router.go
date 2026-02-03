package router

import (
	"agenda-api/internal/config"
	"agenda-api/internal/handlers"
	"agenda-api/internal/middleware"
	"agenda-api/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func Setup(db *sqlx.DB, cfg *config.Config) *gin.Engine {
	gin.SetMode(cfg.GinMode)
	r := gin.Default()

	r.Use(middleware.CORS())

	// Repositories
	userRepo := repository.NewUserRepository(db)
	eventRepo := repository.NewEventRepository(db)
	attendanceRepo := repository.NewAttendanceRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	assignmentRepo := repository.NewAssignmentRepository(db)

	// Handlers
	authHandler := handlers.NewAuthHandler(userRepo, cfg.JWTSecret, cfg.JWTExpirationHours)
	eventHandler := handlers.NewEventHandler(eventRepo, teamRepo, assignmentRepo)
	attendanceHandler := handlers.NewAttendanceHandler(attendanceRepo, eventRepo)
	teamHandler := handlers.NewTeamHandler(teamRepo, userRepo)
	assignmentHandler := handlers.NewAssignmentHandler(assignmentRepo, eventRepo)
	userHandler := handlers.NewUserHandler(userRepo)

	api := r.Group("/api")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.JWTAuth(cfg.JWTSecret), authHandler.Me)
		}

		// Events routes (public)
		events := api.Group("/events")
		{
			events.GET("", eventHandler.GetAll)
			events.GET("/calendar", eventHandler.GetCalendar)
			events.GET("/:id", eventHandler.GetByID)

			// Protected event routes
			events.POST("",
				middleware.JWTAuth(cfg.JWTSecret),
				eventHandler.Create,
			)

			events.PATCH("/:id",
				middleware.JWTAuth(cfg.JWTSecret),
				eventHandler.Update,
			)

			events.DELETE("/:id",
				middleware.JWTAuth(cfg.JWTSecret),
				eventHandler.Delete,
			)

			// Attendance routes
			events.POST("/:id/register",
				middleware.JWTAuth(cfg.JWTSecret),
				attendanceHandler.Register,
			)

			events.DELETE("/:id/register",
				middleware.JWTAuth(cfg.JWTSecret),
				attendanceHandler.Cancel,
			)

			events.GET("/:id/attendees",
				middleware.JWTAuth(cfg.JWTSecret),
				attendanceHandler.GetAttendees,
			)

			// Assignment routes for events
			events.GET("/:id/assignments",
				middleware.JWTAuth(cfg.JWTSecret),
				assignmentHandler.GetByEventID,
			)

			events.POST("/:id/assignments/respond",
				middleware.JWTAuth(cfg.JWTSecret),
				assignmentHandler.Respond,
			)
		}

		// Users routes
		users := api.Group("/users")
		{
			users.GET("/search",
				middleware.JWTAuth(cfg.JWTSecret),
				userHandler.Search,
			)

			users.GET("",
				middleware.JWTAuth(cfg.JWTSecret),
				userHandler.GetAll,
			)

			users.GET("/:id",
				middleware.JWTAuth(cfg.JWTSecret),
				userHandler.GetByID,
			)
		}

		// Teams routes (admin only for management)
		teams := api.Group("/teams")
		{
			teams.GET("",
				middleware.JWTAuth(cfg.JWTSecret),
				teamHandler.GetAll,
			)

			teams.POST("",
				middleware.JWTAuth(cfg.JWTSecret),
				teamHandler.Create,
			)

			teams.GET("/:id",
				middleware.JWTAuth(cfg.JWTSecret),
				teamHandler.GetByID,
			)

			teams.PATCH("/:id",
				middleware.JWTAuth(cfg.JWTSecret),
				teamHandler.Update,
			)

			teams.DELETE("/:id",
				middleware.JWTAuth(cfg.JWTSecret),
				teamHandler.Delete,
			)

			// Team members
			teams.GET("/:id/members",
				middleware.JWTAuth(cfg.JWTSecret),
				teamHandler.GetMembers,
			)

			teams.POST("/:id/members",
				middleware.JWTAuth(cfg.JWTSecret),
				teamHandler.AddMember,
			)

			teams.DELETE("/:id/members/:userId",
				middleware.JWTAuth(cfg.JWTSecret),
				teamHandler.RemoveMember,
			)
		}

		// My routes (user's personal data)
		my := api.Group("/my")
		my.Use(middleware.JWTAuth(cfg.JWTSecret))
		{
			my.GET("/calendar", eventHandler.GetMyCalendar)
			my.GET("/events", eventHandler.GetMyEvents)
			my.GET("/teams", teamHandler.GetMyTeams)
			my.GET("/assignments", assignmentHandler.GetMyAssignments)
			my.GET("/assignments/pending-count", assignmentHandler.GetPendingCount)
			my.GET("/registrations", attendanceHandler.GetMyRegistrations)
		}
	}

	return r
}
