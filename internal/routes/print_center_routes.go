package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kimbasn/printly/internal/controller"
	"github.com/kimbasn/printly/internal/repository"
	"github.com/kimbasn/printly/internal/service"
	"gorm.io/gorm"
)

func RegisterPrintCenterRoutes(rg *gin.RouterGroup, db *gorm.DB, validate *validator.Validate) {
	repo := repository.NewPrintCenterRepository(db)
	svc := service.NewPrintCenterService(repo)
	ctrl := controller.NewPrintCenterController(svc, validate)

	// Publicly accessible print center routes
	publicCenters := rg.Group("/centers")
	{
		publicCenters.GET("/", ctrl.GetAllPublicPrintCenters)
		publicCenters.GET("/:id", ctrl.GetPrintCenterByID)
	}

	// Authenticated routes for print centers.
	// In a real app, this group would have an auth middleware.
	authedCenters := rg.Group("/centers")
	// authedCenters.Use(auth.Middleware()) // Example
	{
		authedCenters.POST("/", ctrl.CreatePrintCenter)
		authedCenters.PUT("/:id", ctrl.UpdatePrintCenter)
	}

	// Admin-specific routes for managing print centers.
	// This group would have an admin-only auth middleware.
	adminCenters := rg.Group("/admin/centers")
	// adminCenters.Use(auth.AdminMiddleware()) // Example
	{
		adminCenters.GET("/", ctrl.GetAllPrintCenters)
		adminCenters.GET("/pending", ctrl.GetPendingPrintCenters)
		adminCenters.PATCH("/:id/status", ctrl.UpdatePrintCenterStatus)
		adminCenters.DELETE("/:id", ctrl.DeletePrintCenter)
	}
}

