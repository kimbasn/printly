package service

type LoginService interface {
	Login(username string, password string) bool
}

type loginService struct {
	authorizeUsername string
	authorizedPassword string
}

func NewLoginService() LoginService {
	return &loginService{
		authorizeUsername: "kimba",
		authorizedPassword: "pwd",
	}
}

func (service *loginService) Login(username string, password string) bool {
	return service.authorizeUsername == username && 
	service.authorizedPassword == password
}