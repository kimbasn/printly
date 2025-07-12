package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kimbasn/printly/internal/entity"
	"github.com/kimbasn/printly/internal/service"
	"github.com/kimbasn/printly/internal/dto"
	"github.com/kimbasn/printly/internal/validators"

	"github.com/go-playground/validator/v10"
)

type UserController interface {
	CreateUser(ctx *gin.Context)
	GetUserByUID(ctx *gin.Context)
	DeleteUserByUID(ctx *gin.Context)
	UpdateUserProfile(ctx *gin.Context)
	GetAllUsers(ctx *gin.Context)
}

type userController struct {
	service service.UserService
	validate *validator.Validate
}

func NewUserController(service service.UserService, validate *validator.Validate) UserController {
	validate.RegisterValidation("is-valid-role", validators.ValidateRole)
	return &userController{
		service: service,
		validate: validate,
	}
}

// CreateUser godoc
// @Summary      Create a new user
// @Description  Registers a new user in the system.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        user  body      dto.CreateUserRequest  true  "User to create"
// @Success      201   {object}  entity.User
// @Failure      400   {object}  dto.ErrorResponse "Invalid input"
// @Failure      409   {object}  dto.ErrorResponse "User already exists"
// @Failure      500   {object}  dto.ErrorResponse "Failed to register user"
// @Router       /users [post]
func (c *userController) CreateUser(ctx *gin.Context) {
	var req dto.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	if err := c.validate.Struct(req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	user := &entity.User{
		UID:         req.UID,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Role:        entity.Role(req.Role),
	}
	
	created, err := c.service.Register(user)
	if err != nil {
		HandleServiceError(ctx, err, "failed to register user")
		return
	}
	ctx.JSON(http.StatusCreated, created)
}

// GetUserByUID godoc
// @Summary      Get a user by UID
// @Description  Retrieves a single user by their unique identifier.
// @Tags         Users
// @Produce      json
// @Param        uid   path      string       true  "User UID"
// @Success      200   {object}  entity.User
// @Failure      404   {object}  dto.ErrorResponse "User not found"
// @Failure      500   {object}  dto.ErrorResponse "Failed to fetch user"
// @Router       /users/{uid} [get]
func (c *userController) GetUserByUID(ctx *gin.Context) {
	uid := ctx.Param("uid")
	user, err := c.service.GetByUID(uid)
	if err != nil {
		HandleServiceError(ctx, err, "failed to fetch user")
		return
	}
	ctx.JSON(http.StatusOK, user)
}

// UpdateUserProfile godoc
// @Summary      Update a user's profile
// @Description  Updates a user's profile information. Only email and phone number can be updated.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        uid   path      string       true  "User UID"
// @Param        user  body      dto.UpdateUserRequest  true  "User data to update"
// @Success      200   {object}  dto.SuccessResponse "Profile updated"
// @Failure      400   {object}  dto.ErrorResponse   "Invalid input"
// @Failure      404   {object}  dto.ErrorResponse   "User not found"
// @Failure      500   {object}  dto.ErrorResponse   "Failed to update profile"
// @Router       /users/{uid} [put]
func (c *userController) UpdateUserProfile(ctx *gin.Context) {
	var req dto.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	if err := c.validate.Struct(req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	user := &entity.User{
		UID:         ctx.Param("uid"),
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
	}
	
	if err := c.service.UpdateProfile(user); err != nil {
		HandleServiceError(ctx, err, "failed to update profile")
		return
	}
	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "profile updated"})
}

// GetAllUsers godoc
// @Summary      Get all users
// @Description  Retrieves a list of all users.
// @Tags         Users
// @Produce      json
// @Success      200   {array}   entity.User
// @Failure      500   {object}  dto.ErrorResponse "Failed to fetch users"
// @Router       /users [get]
func (c *userController) GetAllUsers(ctx *gin.Context) {
	users, err := c.service.GetAll()
	if err != nil {
		HandleServiceError(ctx, err, "failed to fetch users")
		return
	}
	ctx.JSON(http.StatusOK, users)
}

// DeleteUserByUID godoc
// @Summary      Delete a user
// @Description  Deletes a user by their unique identifier.
// @Tags         Users
// @Produce      json
// @Param        uid   path      string  true  "User UID"
// @Success      200   {object}  dto.SuccessResponse "User deleted successfully"
// @Failure      400   {object}  dto.ErrorResponse   "Missing user UID"
// @Failure      404   {object}  dto.ErrorResponse   "User not found"
// @Failure      500   {object}  dto.ErrorResponse   "Failed to delete user"
// @Router       /users/{uid} [delete]
func (c *userController) DeleteUserByUID(ctx *gin.Context) {
	uid := ctx.Param("uid")
	if uid == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "missing user UID"})
		return
	}

	err := c.service.Delete(uid)
	if err != nil {
		HandleServiceError(ctx, err, "failed to delete user")
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "user deleted successfully"})
}


