package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kimbasn/printly/internal/entity"
	"github.com/kimbasn/printly/internal/service"
	"github.com/kimbasn/printly/internal/validators"
	"gopkg.in/go-playground/validator.v9"
)

type UserController interface {
	Register(ctx *gin.Context)
	GetByUID(ctx *gin.Context)
	DeleteByUID(ctx *gin.Context)
	UpdateProfile(ctx *gin.Context)
	FindAll(ctx *gin.Context)
}

type controller struct {
	service service.UserService
}

var validate *validator.Validate

func NewUserController(service service.UserService) UserController {
	validate = validator.New()
	validate.RegisterValidation("is-valid-role", validators.ValidateRole)
	return &controller{
		service: service,
	}
}

func (c *controller) Register(ctx *gin.Context) {
	var u entity.User
	if err := ctx.ShouldBindJSON(&u); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	created, err := c.service.RegisterIfNotExist(&u)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register"})
		return
	}
	ctx.JSON(http.StatusOK, created)
}

func (c *controller) GetByUID(ctx *gin.Context) {
	uid := ctx.Param("uid")
	u, err := c.service.GetByUID(uid)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "fetch failed"})
		return
	}
	if u == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	ctx.JSON(http.StatusOK, u)
}

func (c *controller) UpdateProfile(ctx *gin.Context) {
	var u entity.User
	if err := ctx.ShouldBindJSON(&u); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	u.UID = ctx.Param("uid")
	if err := c.service.UpdateProfile(&u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}

func (c *controller) FindAll(ctx *gin.Context) {
	users := c.service.FindAll()
	ctx.JSON(http.StatusOK, users)
}

func (c *controller) DeleteByUID(ctx *gin.Context) {
	uid := ctx.Param("uid")
	if uid == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "missing user UID"})
		return
	}

	err := c.service.DeleteByUID(uid)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}
