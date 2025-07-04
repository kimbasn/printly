package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kimbasn/printly/internal/controller"
	"github.com/kimbasn/printly/internal/repository"
	"github.com/kimbasn/printly/internal/service"
	"gorm.io/gorm"
)

func RegisterUserRoutes(rg *gin.RouterGroup, db *gorm.DB, validate *validator.Validate) {
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService, validate)

	userGroup := rg.Group("/users")
	{
		userGroup.POST("/", userController.CreateUser)
		userGroup.GET("/", userController.GetAllUsers)
		userGroup.GET("/:uid", userController.GetUserByUID)
		userGroup.PUT("/:uid", userController.UpdateUserProfile)
		userGroup.DELETE("/:uid", userController.DeleteUserByUID)
	}
}
