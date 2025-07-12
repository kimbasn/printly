package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kimbasn/printly/internal/controller"
	"github.com/kimbasn/printly/internal/repository"
	"github.com/kimbasn/printly/internal/service"
	"gorm.io/gorm"
)

func RegisterOrderRoutes(rg *gin.RouterGroup, db *gorm.DB, validate *validator.Validate) {
	// Repositories
	orderRepo := repository.NewOrderRepository(db)
	printCenterRepo := repository.NewPrintCenterRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Service & Controller
	orderService := service.NewOrderService(orderRepo, printCenterRepo, userRepo)
	orderController := controller.NewOrderController(orderService, validate)

	// Middleware

	// Public routers
	rg.GET("/orders/status/:code", orderController.GetOrderByCode)

	// Authenticated routes
	authed := rg.Group("/")
	
	// Any authenticated user
	authed.POST("/centers/:id/orders", orderController.CreateOrder)
	
	// A manager or admin
	authed.GET("/centers/:id/orders", orderController.GetOrdersForCenter)
	authed.PATCH("/orders/:id/status", orderController.UpdateOrderStatus)


	// Admin-specific routes
	admin := rg.Group("/admin")
	orders := admin.Group("/orders")
	orders.GET("/", orderController.GetAllOrders)
	orders.GET("/:id", orderController.GetOrderByID)
	orders.DELETE("/:id", orderController.DeleteOrder)
}