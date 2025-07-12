package middlewares

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/kimbasn/printly/internal/dto"
	"github.com/kimbasn/printly/internal/entity"
	"github.com/kimbasn/printly/internal/repository"
	"gorm.io/gorm"
)

func BasicAuth() gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		"kimba": "pwd",
		"sabi":	"pwd",
	})
}

func AuthenticationMiddleware(app *firebase.App, db *gorm.DB) gin.HandlerFunc {
	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting auth client: %v\n", err)
	}

	userRepo := repository.NewUserRepository(db)

	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Authorization header is required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Authorization header format must be Bearer {toke}"})
			return
		}

		idToken := parts[1]
		token, err := client.VerifyIDToken(ctx, idToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid or expired token"})
			return
		}

		// TOken is valid, now get user from DB to check role
		user, err := userRepo.FindByUID(token.UID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// User exists in Firebase but not in our DB
				// THis must not happpen because user creation process is as is:
				// Create user in firebase then create in db in transaction manner( both must succeed)
				ctx.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{Error: "User not found in the system"})
			} else {
				log.Printf("Database error fetching user by UID %s: %v", token.UID, err)
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Error fetching user information"})
			}
			return
		}

		// Set user info in context for downstream handlers
		ctx.Set("userUID", user.UID)
		ctx.Set("userRole", user.Role)
		ctx.Set("user", user) // Storing the whole user, it can be useful

		ctx.Next()
	}
}

func RoleMiddleware(allowedRoles ...entity.Role) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		v, exists := ctx.Get("userRole")
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "User role not found in context"})
			return
		}

		userRole, ok := v.(entity.Role)
		if !ok {
			// This indicates a programming error, where a non-entity.Role value was set for "userRole"
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Invalid user role type in context"})
			return
		}

		// First, check if the user's role is a valid one.
		if !userRole.IsValid() {
			ctx.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{Error: "User has an unrecognized role"})
			return
		}

		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				ctx.Next()
				return
			}
		}
		ctx.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{Error: "You do not have permission to access this resource"})
	}
}
