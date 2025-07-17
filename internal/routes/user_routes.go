package routes

import (
	"context"
	"log"

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

func RegisterUserRoutes(rg *gin.RouterGroup, db *gorm.DB, validate *validator.Validate, fbApp *firebase.App) {
	userRepo := repository.NewUserRepository(db)

	fbAuthClient, err := fbApp.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting firebase auth client: %v", err)
	}
	
	userService := service.NewUserService(userRepo, fbAuthClient)

	userController := controller.NewUserController(userService, validate)

	// Authenticated routes for users to manage their own profile.
	// Example: GET /api/v1/users/me
	me := rg.Group("users/me")
	me.Use(middlewares.AuthenticationMiddleware(fbApp, db))
	{
		me.GET("/", userController.GetMyProfile)
		me.PATCH("/", userController.UpdateMyProfile)
		me.DELETE("/", userController.DeleteMyProfile)
	}

	// Admin routes for managing all users.
	adminUsers := rg.Group("/admin/users")
	adminUsers.Use(middlewares.AuthenticationMiddleware(fbApp, db),
		middlewares.RoleMiddleware(entity.RoleAdmin))
	{
		adminUsers.POST("/", userController.CreateUser)
		adminUsers.GET("/", userController.GetAllUsers)
		adminUsers.GET("/:uid", userController.GetUserByUID)
		adminUsers.PUT("/:uid", userController.UpdateUserProfile)
		adminUsers.PATCH("/:uid/role", userController.UpdateUserRole)
		adminUsers.DELETE("/:uid", userController.DeleteUserByUID)
	}
}
