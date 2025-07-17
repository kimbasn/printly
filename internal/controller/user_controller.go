package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kimbasn/printly/internal/dto"
	"github.com/kimbasn/printly/internal/entity"
	"github.com/kimbasn/printly/internal/service"
	"github.com/kimbasn/printly/internal/validators"

	"github.com/go-playground/validator/v10"
)

type UserController interface {
	CreateUser(ctx *gin.Context)
	GetUserByUID(ctx *gin.Context)
	DeleteUserByUID(ctx *gin.Context)
	UpdateUserProfile(ctx *gin.Context)
	GetAllUsers(ctx *gin.Context)
	GetMyProfile(ctx *gin.Context)
	UpdateMyProfile(ctx *gin.Context)
	DeleteMyProfile(ctx *gin.Context)
	UpdateUserRole(ctx *gin.Context)
}

type userController struct {
	service  service.UserService
	validate *validator.Validate
}

func NewUserController(service service.UserService, validate *validator.Validate) UserController {
	validate.RegisterValidation("is-valid-role", validators.ValidateRole)
	return &userController{
		service:  service,
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
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Role:      entity.RoleUser, // Default role
		Disabled:  false,
	}

	createdUser, err := c.service.Register(user, req.Password)
	if err != nil {
		HandleServiceError(ctx, err, "failed to register user")
		return
	}
	response := dto.UserResponse{
		UID:       createdUser.UID,
		FirstName: createdUser.FirstName,
		LastName:  createdUser.LastName,
		Email:     createdUser.Email,
		Role:      createdUser.Role,
		Disabled:  createdUser.Disabled,
	}
	ctx.JSON(http.StatusCreated, response)
}

// GetUserByUID godoc
// @Summary      Get a user by UID
// @Description  Retrieves a single user by their unique identifier.
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
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
// @Security     BearerAuth
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

	uid := ctx.Param("uid")
	updates := make(map[string]any)
	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Disabled {
		updates["disabled"] = req.Disabled
	}

	if err := c.service.UpdateProfile(uid, updates); err != nil {
		HandleServiceError(ctx, err, "failed to update profile")
		return
	}
	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "profile updated successfully"})
}

// GetAllUsers godoc
// @Summary      Get all users
// @Description  Retrieves a list of all users.
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
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
// @Security     BearerAuth
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

// GetMyProfile godoc
// @Summary      Get current user's profile
// @Description  Retrieves the profile of the currently authenticated user.
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  entity.User
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized"
// @Failure      404  {object}  dto.ErrorResponse "User not found"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Router       /users/me [get]
func (c *userController) GetMyProfile(ctx *gin.Context) {
	// The userUID is set by the AuthenticationMiddleware.
	userUID, exists := ctx.Get("userUID")
	if !exists {
		// This case should ideally not be reached if the middleware is applied correctly.
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "user UID not found in context"})
		return
	}

	user, err := c.service.GetByUID(userUID.(string))
	if err != nil {
		HandleServiceError(ctx, err, "failed to fetch user profile")
		return
	}
	ctx.JSON(http.StatusOK, user)
}

// UpdateMyProfile godoc
// @Summary      Update current user's profile
// @Description  Allows the currently authenticated user to update their profile information.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user  body      dto.UpdateUserRequest  true  "Profile data to update"
// @Success      200   {object}  dto.SuccessResponse "Profile updated successfully"
// @Failure      400   {object}  dto.ErrorResponse   "Invalid input"
// @Failure      401   {object}  dto.ErrorResponse   "Unauthorized"
// @Failure      404   {object}  dto.ErrorResponse   "User not found"
// @Failure      500   {object}  dto.ErrorResponse   "Failed to update profile"
// @Router       /users/me [patch]
func (c *userController) UpdateMyProfile(ctx *gin.Context) {
	// The userUID is set by the AuthenticationMiddleware.
	userUID, exists := ctx.Get("userUID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "user UID not found in context"})
		return
	}

	var req dto.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Build a map of fields to update to avoid overwriting with zero values.
	updates := make(map[string]any)
	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Disabled {
		updates["disabled"] = req.Disabled
	}

	if len(updates) == 0 {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "no fields to update"})
		return
	}

	if err := c.service.UpdateProfile(userUID.(string), updates); err != nil {
		HandleServiceError(ctx, err, "failed to update profile")
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "profile updated successfully"})
}

// DeleteMyProfile godoc
// @Summary      Delete current user's account
// @Description  Permanently deletes the account of the currently authenticated user from the system and Firebase.
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.SuccessResponse "Account deleted successfully"
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized"
// @Failure      404  {object}  dto.ErrorResponse "User not found"
// @Failure      500  {object}  dto.ErrorResponse "Failed to delete account"
// @Router       /users/me [delete]
func (c *userController) DeleteMyProfile(ctx *gin.Context) {
	userUID, exists := ctx.Get("userUID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "user UID not found in context"})
		return
	}

	uid := userUID.(string)
	if err := c.service.Delete(uid); err != nil {
		HandleServiceError(ctx, err, "failed to delete account")
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "account deleted successfully"})
}

// UpdateUserRole godoc
// @Summary      Update a user's role
// @Description  Sets the role for a specific user. Requires admin privileges.
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        uid   path      string                      true  "User UID"
// @Param        role  body      dto.UpdateUserRoleRequest   true  "New role for the user"
// @Success      200   {object}  dto.SuccessResponse "Role updated successfully"
// @Failure      400   {object}  dto.ErrorResponse   "Invalid input or role"
// @Failure      404   {object}  dto.ErrorResponse   "User not found"
// @Failure      500   {object}  dto.ErrorResponse   "Failed to update role"
// @Router       /admin/users/{uid}/role [patch]
func (c *userController) UpdateUserRole(ctx *gin.Context) {
	uid := ctx.Param("uid")

	var req dto.UpdateUserRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid role specified"})
		return
	}

	if err := c.service.UpdateRole(uid, entity.Role(req.Role)); err != nil {
		HandleServiceError(ctx, err, "failed to update user role")
		return
	}
	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "role updated successfully"})
}
