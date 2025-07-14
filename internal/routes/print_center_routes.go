package routes

import (
	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kimbasn/printly/internal/controller"
	"github.com/kimbasn/printly/internal/entity"
	"github.com/kimbasn/printly/internal/middlewares"
	"github.com/kimbasn/printly/internal/repository"
	"github.com/kimbasn/printly/internal/service"
	"gorm.io/gorm"
)

func RegisterPrintCenterRoutes(rg *gin.RouterGroup, db *gorm.DB, validate *validator.Validate, fbApp *firebase.App) {
	repo := repository.NewPrintCenterRepository(db)
	svc := service.NewPrintCenterService(repo)
	printCenterController := controller.NewPrintCenterController(svc, validate)

	// Publicly accessible print center routes
	publicCenters := rg.Group("/centers")
	{
		publicCenters.GET("/", printCenterController.GetAllPublicPrintCenters)
		publicCenters.GET("/:id", printCenterController.GetPrintCenterByID)
	}

	// Authenticated routes for any logged-in user.
	authed := rg.Group("/")
	authed.Use(middlewares.AuthenticationMiddleware(fbApp, db))
	{
		authed.POST("/centers", printCenterController.CreatePrintCenter) //  any authenticated user
		authed.PUT("/centers/:id", middlewares.RoleMiddleware(entity.RoleManager, entity.RoleAdmin), printCenterController.UpdatePrintCenter)
	}

	// Admin-specific routes for managing print centers.
	admin := rg.Group("/admin")
	admin.Use(middlewares.AuthenticationMiddleware(fbApp, db),
		middlewares.RoleMiddleware(entity.RoleAdmin))
	{
		admin.GET("/centers", printCenterController.GetAllPrintCenters)
		admin.GET("/centers/pending", printCenterController.GetPendingPrintCenters)
		admin.PATCH("centers/:id/status", printCenterController.UpdatePrintCenterStatus)
		admin.DELETE("centers/:id", printCenterController.DeletePrintCenter)
	}
}
