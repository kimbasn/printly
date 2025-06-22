package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kimbasn/printly/internal/controller"
	"github.com/kimbasn/printly/internal/db"
	"github.com/kimbasn/printly/internal/repository"
	"github.com/kimbasn/printly/internal/service"
)

func RegisterUserRoutes(rg *gin.RouterGroup) {
	userRepo := repository.NewUserRepository(db.DB)
	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService)

	userGroup := rg.Group("/users")
	{
		userGroup.POST("/", userController.Register)
		userGroup.GET("/", userController.FindAll)
		userGroup.GET("/:uid", userController.GetByUID)
		userGroup.PUT("/:uid", userController.UpdateProfile)
		userGroup.DELETE("/:uid", userController.DeleteByUID)
	}
}
