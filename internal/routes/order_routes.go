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
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func RegisterOrderRoutes(rg *gin.RouterGroup, db *gorm.DB, validate *validator.Validate, fbApp *firebase.App, logger *zap.Logger, storageService service.StorageService) {
	// Repositories
	orderRepo := repository.NewOrderRepository(db)
	printCenterRepo := repository.NewPrintCenterRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Service & Controller
	orderService := service.NewOrderService(orderRepo,
		printCenterRepo,
		userRepo,
		logger)
	orderController := controller.NewOrderController(orderService,
		storageService,
		validate,
		logger)

	// Public route for checking order status
	rg.GET("/orders/status/:code", orderController.GetOrderByCode)

	// Any authenticated user
	authed := rg.Group("/")
	authed.Use(middlewares.AuthenticationMiddleware(fbApp, db))
	{
		// any authenticated user
		authed.POST("/centers/:id/orders", orderController.CreateOrder)

		// manager + admin
		authed.GET("centers/:id/orders", middlewares.RoleMiddleware(entity.RoleManager, entity.RoleAdmin), orderController.GetOrdersForCenter)
		authed.PATCH("/orders/:id/status", middlewares.RoleMiddleware(entity.RoleManager, entity.RoleAdmin), orderController.UpdateOrderStatus)
	}

	// Admin-specific routes
	admin := rg.Group("/admin")
	admin.Use(middlewares.AuthenticationMiddleware(fbApp, db),
		middlewares.RoleMiddleware(entity.RoleAdmin))
	admin.GET("/orders", orderController.GetAllOrders)
	admin.GET("orders/:id", orderController.GetOrderByID)
	admin.DELETE("orders/:id", orderController.DeleteOrder)
}
