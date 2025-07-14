package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kimbasn/printly/internal/dto"
	"github.com/kimbasn/printly/internal/service"
)

type LoginController interface {
	Login(ctx *gin.Context)
}

type loginController struct {
	loginService service.LoginService
	jwtService   service.JWTService
}

func NewLoginController(loginService service.LoginService,
	jwtService service.JWTService) LoginController {
	return &loginController{
		loginService: loginService,
		jwtService:   jwtService,
	}
}

func (c *loginController) Login(ctx *gin.Context) {
	var credentials dto.Credentials
	err := ctx.ShouldBind(&credentials)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid request"})
		return
	}
	isAuthenticated := c.loginService.Login(credentials.Phone_number,
		credentials.Password)
	if isAuthenticated {
		token := c.jwtService.GenerateToken(credentials.Phone_number, false)
		ctx.JSON(http.StatusOK, dto.JWT{
			Token: token,
		})
		return
	}
	ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized"})
}
